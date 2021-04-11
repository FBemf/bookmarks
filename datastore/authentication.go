package datastore

import (
	"crypto"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const AUTH_COOKIE_NAME = "bookmark_auth"
const AUTH_COOKIE_SIZE = 32
const AUTH_COOKIE_TTL = 30 * 24 * 60 * 60 * time.Second // 30 days in seconds
const SALT_SIZE = 16

func (ds *Datastore) GetSession(cookie string) (string, bool, error) {
	var user int64
	var timestamp time.Time
	err := ds.db.QueryRow(`select user, timestamp from session where cookie = ?`, cookie).Scan(&user, &timestamp)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("getting session: %w", err)
	}

	if time.Now().UTC().Sub(timestamp) > AUTH_COOKIE_TTL {
		_, err := ds.db.Exec(`delete from session where cookie = ?`, cookie)
		if err != nil {
			return "", false, fmt.Errorf("deleting expired cookie: %w", err)
		}
		return "", false, fmt.Errorf("expired cookie")
	}

	var username string
	err = ds.db.QueryRow(`select username from user where id = ?`, user).Scan(&username)
	if err != nil {
		return "", false, fmt.Errorf("getting username: %w", err)
	}
	return username, true, nil
}

func (ds *Datastore) AddUser(username, password string) error {
	salt, err := createSalt()
	if err != nil {
		return fmt.Errorf("creating salt: %w", err)
	}
	hash := hashPassword(password, salt)
	_, err = ds.db.Exec(`insert into user (username, password_hash, salt) values (?, ?, ?)`, username, hash, salt)
	if err != nil {
		return fmt.Errorf("inserting new user: %w", err)
	}
	return nil
}

func (ds *Datastore) ChangeUserPassword(username, password string) error {
	salt, err := createSalt()
	if err != nil {
		return fmt.Errorf("creating salt: %w", err)
	}
	hash := hashPassword(password, salt)
	_, err = ds.db.Exec(`update user set password_hash = ?, salt = ? where username = ?`, hash, salt, username)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func createSalt() (string, error) {
	saltBytes := make([]byte, SALT_SIZE)
	_, err := rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)
	if err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	return salt, nil
}

func hashPassword(password, salt string) string {
	hasher := crypto.SHA512.New()
	hashBytes := hasher.Sum([]byte(password + salt))
	hash := hex.EncodeToString(hashBytes)
	return hash
}

func (ds *Datastore) ListUsers() ([]string, error) {
	rows, err := ds.db.Query(`select username from user`)
	if err != nil {
		return nil, err
	}
	users := make([]string, 0)
	for rows.Next() {
		var user string
		rows.Scan(&user)
		users = append(users, user)
	}
	return users, nil
}

func (ds *Datastore) UserExists(username string) (int64, bool, error) {
	var userId int64
	err := ds.db.QueryRow(`select id from user where username = ?`, username).Scan(&userId)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return userId, true, nil
}

func (ds *Datastore) AuthenticateUser(username, password string) (int64, bool, error) {
	var userId int64
	var stored_hash string
	var salt string
	err := ds.db.QueryRow(`select id, password_hash, salt from user where username = ?`, username).
		Scan(&userId, &stored_hash, &salt)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("finding user: %w", err)
	}

	new_hash := hashPassword(password, salt)

	if stored_hash == new_hash {
		return userId, true, nil
	} else {
		return 0, false, nil
	}
}

func (ds *Datastore) RemoveUser(username string) error {
	_, err := ds.db.Exec(`delete from user where username = ?`, username)
	return err
}

func (ds *Datastore) CreateSession(user int64) (http.Cookie, error) {
	cookieBytes := make([]byte, AUTH_COOKIE_SIZE)
	_, err := rand.Read(cookieBytes)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("generating cookie: %w", err)
	}
	cookie := hex.EncodeToString(cookieBytes)
	timestamp := time.Now().UTC()
	_, err = ds.db.Exec(`insert into session (user, timestamp, cookie) values (?, ?, ?)`, user, timestamp, cookie)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("inserting session: %w", err)
	}
	return http.Cookie{
		Name:     AUTH_COOKIE_NAME,
		Value:    cookie,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}, nil
}

func (ds *Datastore) CleanUpCookies(ttl time.Duration) error {
	oldestAllowed := time.Now().UTC().Add(-ttl)
	_, err := ds.db.Exec(`delete from session where timestamp < ?`, oldestAllowed)
	return err
}
