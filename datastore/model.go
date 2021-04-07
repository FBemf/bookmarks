package datastore

import (
	"database/sql"
	"fmt"
	"time"
)

type Datastore struct {
	db *sql.DB
}

type Bookmark struct {
	Id          int
	Name        string
	Date        time.Time
	Url         string
	Description string
}

func Connect(file string) (Datastore, error) {
	address := fmt.Sprintf("file:%s?_foreign_keys=1", file)
	db, err := sql.Open("sqlite3", address)
	if err != nil {
		return Datastore{}, fmt.Errorf("Unable to open sqlite3 connection: %w", err)
	}
	return Datastore{db}, nil
}

func (ds *Datastore) NewBookmark(name, url, description string) error {
	date := time.Now().UTC()
	_, err := ds.db.Exec(`insert into bookmark (name, date, url, description) values (?, ?, ?, ?)`, name, date, url, description)
	if err != nil {
		return fmt.Errorf("inserting new bookmark: %w", err)
	}
	return nil
}

func (ds *Datastore) RecentBookmarks(number uint) ([]Bookmark, error) {
	result := make([]Bookmark, 0, number)
	rows, err := ds.db.Query(`select * from bookmark order by date desc limit ?`, number)
	if err != nil {
		return result, fmt.Errorf("fetching bookmarks: %w", err)
	}
	for rows.Next() {
		var b Bookmark
		err = rows.Scan(&b.Id, &b.Name, &b.Url, &b.Date, &b.Description)
		if err != nil {
			return result, fmt.Errorf("scanning bookmark: %w", err)
		}
		result = append(result, b)
	}
	return result, nil
}
