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

const AuthCookieName = "bookmark_auth"
const AuthCookieTtl = 30 * 24 * 60 * 60 * time.Second // 30 days in seconds
const authCookieSize = 32
const saltSize = 16
const csrfTokenSize = 32

func (ds *Datastore) AddUser(username, password string) error {
	salt, err := randomHex(saltSize)
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
	salt, err := randomHex(saltSize)
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

func randomHex(bytes int) (string, error) {
	outputBytes := make([]byte, bytes)
	_, err := rand.Read(outputBytes)
	if err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	output := hex.EncodeToString(outputBytes)
	return output, nil
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
	cookie, err := randomHex(authCookieSize)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("generating cookie: %w", err)
	}
	csrf, err := randomHex(csrfTokenSize)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("generating csrf token: %w", err)
	}
	timestamp := time.Now().UTC()
	_, err = ds.db.Exec(`insert into session (user, timestamp, cookie, csrf) values (?, ?, ?, ?)`, user, timestamp, cookie, csrf)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("inserting session: %w", err)
	}
	return http.Cookie{
		Name:     AuthCookieName,
		Value:    cookie,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}, nil
}

type Session struct {
	UserId    int64
	Username  string
	CsrfToken string
}

func (ds *Datastore) GetSession(cookie string) (Session, bool, error) {
	var user int64
	var timestamp time.Time
	var csrf string
	err := ds.db.QueryRow(`select user, timestamp, csrf from session where cookie = ?`, cookie).Scan(&user, &timestamp, &csrf)
	if err == sql.ErrNoRows {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, fmt.Errorf("getting session: %w", err)
	}

	if time.Now().UTC().Sub(timestamp) > AuthCookieTtl {
		_, err := ds.db.Exec(`delete from session where cookie = ?`, cookie)
		if err != nil {
			return Session{}, false, fmt.Errorf("deleting expired cookie: %w", err)
		}
		return Session{}, false, fmt.Errorf("expired cookie")
	}

	var username string
	err = ds.db.QueryRow(`select username from user where id = ?`, user).Scan(&username)
	if err != nil {
		return Session{}, false, fmt.Errorf("getting username: %w", err)
	}
	return Session{user, username, csrf}, true, nil
}

func (ds *Datastore) CleanUpSessions(ttl time.Duration) error {
	oldestAllowed := time.Now().UTC().Add(-ttl)
	_, err := ds.db.Exec(`delete from session where timestamp < ?`, oldestAllowed)
	return err
}
