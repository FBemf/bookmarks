package datastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Datastore struct {
	db *sql.DB
}

type Bookmark struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	Url         string    `json:"url"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
}

type QueryInfo struct {
	Reverse bool
	Search  string
	Number  uint
	Offset  uint
	Tags    []string
}

func NewQueryInfo(pageSize uint) QueryInfo {
	return QueryInfo{
		Reverse: false,
		Search:  "",
		Number:  pageSize,
		Offset:  0,
		Tags:    make([]string, 0),
	}
}

func Connect(file string) (Datastore, error) {
	address := fmt.Sprintf("file:%s?_foreign_keys=1", file)
	db, err := sql.Open("sqlite3", address)
	if err != nil {
		return Datastore{}, fmt.Errorf("unable to open sqlite3 connection: %w", err)
	}
	return Datastore{db}, nil
}

func (ds *Datastore) GetBookmark(id int64) (Bookmark, error) {
	var result Bookmark
	err := ds.db.QueryRow(`select * from bookmark where id=?`, id).
		Scan(&result.Id, &result.Name, &result.Url, &result.Date, &result.Description)
	if err != nil {
		return result, fmt.Errorf("retrieving bookmark: %w", err)
	}
	result.Tags, err = ds.getTags(id)
	if err != nil {
		return result, fmt.Errorf("retrieving tags: %w", err)
	}
	return result, nil
}

func (ds *Datastore) CreateBookmark(name, url, description string, tags []string) error {
	date := time.Now().UTC()
	ctx, stop := context.WithCancel(context.Background())
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		stop()
		return fmt.Errorf("beginning transaction: %w", err)
	}

	result, err := tx.Exec(
		`insert into bookmark (name, date, url, description) values (?, ?, ?, ?)`,
		name, date, url, description)
	if err != nil {
		stop()
		return fmt.Errorf("inserting bookmark: %w", err)
	}

	bookmarkId, err := result.LastInsertId()
	if err != nil {
		stop()
		return fmt.Errorf("getting bookmark id: %w", err)
	}

	err = setBookmarkTags(bookmarkId, tags, tx)
	if err != nil {
		stop()
		return fmt.Errorf("setting tags: %w", err)
	}

	err = tx.Commit()
	stop()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return err
}

func setBookmarkTags(bookmarkId int64, tags []string, tx *sql.Tx) error {
	lowerTags := stringsToLower(tags)
	for _, tag := range lowerTags {
		var exists int
		err := tx.QueryRow(`select count(*) from tag where name = ?`, tag).Scan(&exists)
		if err != nil {
			return fmt.Errorf("finding whether tag %s exists: %w", tag, err)
		}

		var tagId int64
		if exists == 0 {
			result, err := tx.Exec(`insert or ignore into tag (name) values (?)`, tag)
			if err != nil {
				return fmt.Errorf("creating tag %s: %w", tag, err)
			}

			tagId, err = result.LastInsertId()
			if err != nil {
				return fmt.Errorf("getting tag %s id: %w", tag, err)
			}
		} else {
			err = tx.QueryRow(`select id from tag where name = ?`, tag).Scan(&tagId)
			if err != nil {
				return fmt.Errorf("getting id of tag %s: %w", tag, err)
			}
		}

		_, err = tx.Exec(`insert or ignore into tag_bookmark (tag, bookmark) values (?, ?)`, tagId, bookmarkId)
		if err != nil {
			return fmt.Errorf("tagging bookmark %d with tag %s: %w", bookmarkId, tag, err)
		}
	}

	// clear bookmarks we didn't just insert
	query := fmt.Sprintf(
		`delete from tag_bookmark where bookmark = ? and tag not in (select id from tag where name in (%s))`,
		quoteStrings(lowerTags),
	)
	_, err := tx.Exec(query, bookmarkId)
	if err != nil {
		return fmt.Errorf("deleting extra tags: %w", err)
	}

	return nil
}

func (ds *Datastore) UpdateBookmark(id int64, name, url, description string, tags []string) error {
	ctx, stop := context.WithCancel(context.Background())
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		stop()
		return fmt.Errorf("beginning transaction: %w", err)
	}
	_, err = tx.Exec(`update bookmark set name=?, url=?, description=? where id=?`, name, url, description, id)
	if err != nil {
		stop()
		return fmt.Errorf("updating bookmark: %w", err)
	}
	err = setBookmarkTags(id, tags, tx)
	if err != nil {
		stop()
		return fmt.Errorf("setting tags: %w", err)
	}
	err = tx.Commit()
	stop()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	err = ds.deleteDanglingTags()
	if err != nil {
		return fmt.Errorf("deleting dangling tags: %w", err)
	}
	return nil
}

func (ds *Datastore) DeleteBookmark(id int64) error {
	_, err := ds.db.Exec(`delete from bookmark where id=?`, id)
	if err != nil {
		return fmt.Errorf("deleting bookmark: %w", err)
	}
	err = ds.deleteDanglingTags()
	if err != nil {
		return fmt.Errorf("deleting dangling tags: %w", err)
	}
	return nil
}

func (ds *Datastore) GetBookmarks(info QueryInfo) ([]Bookmark, error) {
	result := make([]Bookmark, 0, info.Number)
	var order string
	if info.Reverse {
		order = "asc"
	} else {
		order = "desc"
	}
	tags := stringsToLower(info.Tags)

	var rows *sql.Rows
	var err error
	if len(tags) == 0 {
		if info.Search == "" {
			query := fmt.Sprintf(`select * from bookmark order by date %s limit ? offset ?`, order)
			rows, err = ds.db.Query(query, info.Number, info.Offset)
			if err != nil {
				return result, fmt.Errorf("fetching bookmarks: %w", err)
			}
		} else {
			query := fmt.Sprintf(`select * from bookmark
				where bookmark.name like $1 or url like $1 or description like $1
				order by date %s limit $2 offset $3`, order)
			rows, err = ds.db.Query(query, "%"+info.Search+"%", info.Number, info.Offset)
			if err != nil {
				return result, fmt.Errorf("fetching bookmarks: %w", err)
			}
		}
	} else {
		query := fmt.Sprintf(`select bookmark.id, bookmark.name, url, date, description from bookmark
			join (
				select * from tag_bookmark
				join tag on tag.id = tag_bookmark.tag
				where tag.name in (%s)
				group by bookmark
				having count(distinct tag.id) = %d
			) as t on bookmark.id = t.bookmark
			where bookmark.name like $1 or url like $1 or description like $1
			order by date %s limit $2 offset $3`,
			quoteStrings(tags), len(tags), order)
		pattern := "%" + info.Search + "%"
		rows, err = ds.db.Query(query, pattern, info.Number, info.Offset)
		if err != nil {
			return result, fmt.Errorf("fetching bookmarks: %w", err)
		}
	}
	for rows.Next() {
		var b Bookmark
		err = rows.Scan(&b.Id, &b.Name, &b.Url, &b.Date, &b.Description)
		if err != nil {
			return result, fmt.Errorf("scanning bookmark: %w", err)
		}

		b.Tags, err = ds.getTags(b.Id)
		if err != nil {
			return result, fmt.Errorf("getting tags for bookmark %d: %w", b.Id, err)
		}

		result = append(result, b)
	}
	return result, nil
}

func (ds *Datastore) getTags(bookmarkId int64) ([]string, error) {
	rows, err := ds.db.Query(
		`select name from tag_bookmark inner join tag on tag.id = tag_bookmark.tag where tag_bookmark.bookmark = ?`,
		bookmarkId)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	tags := make([]string, 0, 10)
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return tags, fmt.Errorf("scanning bookmark: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (ds *Datastore) GetNumBookmarks(info QueryInfo) (uint, error) {
	var count uint
	var err error
	if info.Search == "" {
		err = ds.db.QueryRow(`select count(*) from bookmark`).Scan(&count)
	} else {
		err = ds.db.QueryRow(`select count(*) from bookmark
			where name like $1 or url like $1 or description like $1`, "%"+info.Search+"%").
			Scan(&count)
	}
	return count, err
}

func (ds *Datastore) Export() ([]byte, error) {
	n, err := ds.GetNumBookmarks(NewQueryInfo(0))
	if err != nil {
		return nil, fmt.Errorf("retrieving number of bookmarks: %w", err)
	}
	bookmarks, err := ds.GetBookmarks(NewQueryInfo(n))
	if err != nil {
		return nil, fmt.Errorf("retrieving bookmarks: %w", err)
	}
	data, err := json.Marshal(bookmarks)
	if err != nil {
		return nil, fmt.Errorf("marshalling json: %w", err)
	}
	return data, nil
}

func quoteStrings(value []string) string {
	escaped := make([]string, 0, len(value))
	for _, s := range value {
		escaped = append(escaped, strings.Replace(s, "'", "''", -1))
	}
	joined := "'" + strings.Join(escaped, "', '") + "'"
	return joined
}

func (ds *Datastore) deleteDanglingTags() error {
	_, err := ds.db.Exec(`delete from tag where (select count(*) from tag_bookmark where tag = id) = 0`)
	return err
}

func stringsToLower(input []string) []string {
	output := make([]string, 0, len(input))
	for _, s := range input {
		output = append(output, strings.ToLower(s))
	}
	return output
}
