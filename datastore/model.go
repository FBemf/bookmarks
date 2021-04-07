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

func (ds *Datastore) GetBookmark(id int) (Bookmark, error) {
	var result Bookmark
	err := ds.db.QueryRow(`select * from bookmark where id=?`, id).
		Scan(&result.Id, &result.Name, &result.Url, &result.Date, &result.Description)
	return result, err
}

func (ds *Datastore) CreateBookmark(name, url, description string) error {
	date := time.Now().UTC()
	_, err := ds.db.Exec(
		`insert into bookmark (name, date, url, description) values (?, ?, ?, ?)`,
		name, date, url, description)
	return err
}

func (ds *Datastore) UpdateBookmark(id int, name, url, description string) error {
	_, err := ds.db.Exec(`update bookmark set name=?, url=?, description=? where id=?`, name, url, description, id)
	return err
}

func (ds *Datastore) DeleteBookmark(id int) error {
	_, err := ds.db.Exec(`delete from bookmark where id=?`, id)
	return err
}

func (ds *Datastore) GetRecentBookmarks(number, offset uint) ([]Bookmark, error) {
	result := make([]Bookmark, 0, number)
	rows, err := ds.db.Query(`select * from bookmark order by date desc limit ? offset ?`, number, offset)
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

func (ds *Datastore) GetNumBookmarks() (uint, error) {
	var count uint
	err := ds.db.QueryRow(`select count(*) from bookmark count`).Scan(&count)
	return count, err
}
