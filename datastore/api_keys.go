package datastore

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

const API_KEY_SIZE = 32

type ApiKey struct {
	Id   int64
	Name string
	Key  string
}

func (ds *Datastore) CreateKey(name string) error {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return fmt.Errorf("generating cookie: %w", err)
	}
	key := hex.EncodeToString(keyBytes)
	timestamp := time.Now().UTC()
	_, err = ds.db.Exec(`insert into api_key (name, key, timestamp) values (?, ?, ?)`, name, key, timestamp)
	if err != nil {
		return fmt.Errorf("inserting key: %w", err)
	}
	return nil
}

func (ds *Datastore) ListKeys() ([]ApiKey, error) {
	rows, err := ds.db.Query(`select id, name, key from api_key order by timestamp desc`)
	if err != nil {
		return nil, fmt.Errorf("getting rows: %w", err)
	}
	keys := make([]ApiKey, 0)
	for rows.Next() {
		var key ApiKey
		err = rows.Scan(&key.Id, &key.Name, &key.Key)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (ds *Datastore) DeleteKey(key int64) error {
	_, err := ds.db.Exec(`delete from api_key where id = ?`, key)
	return err
}

func (ds *Datastore) CheckKey(key string) (string, bool, error) {
	var name string
	err := ds.db.QueryRow(`select name from api_key where key = ?`, key).Scan(&name)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("inserting key: %w", err)
	}
	return name, true, nil
}
