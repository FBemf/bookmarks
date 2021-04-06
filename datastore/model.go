package datastore

import (
	"database/sql"
	"fmt"
)

type Datastore struct {
	db *sql.DB
}

func Connect(file string) (Datastore, error) {
	address := fmt.Sprintf("file:%s?_foreign_keys=1", file)
	db, err := sql.Open("sqlite3", address)
	if err != nil {
		return Datastore{}, fmt.Errorf("Unable to open sqlite3 connection: %w", err)
	}
	return Datastore{db}, nil
}
