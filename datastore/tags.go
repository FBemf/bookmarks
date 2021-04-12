package datastore

import (
	"database/sql"
	"fmt"
)

type Tag struct {
	Name  string
	Count int64
}

func (ds *Datastore) GetTags() ([]Tag, error) {
	rows, err := ds.db.Query(
		`select name, count(bookmark) from tag
		join tag_bookmark on tag.id = tag_bookmark.tag
		group by name order by name asc`)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		err = rows.Scan(&tag.Name, &tag.Count)
		if err != nil {
			return tags, fmt.Errorf("scanning bookmark: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (ds *Datastore) getBookmarkTags(bookmarkId int64) ([]string, error) {
	rows, err := ds.db.Query(
		`select name from tag_bookmark inner join tag on tag.id = tag_bookmark.tag where tag_bookmark.bookmark = ?`,
		bookmarkId)
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	tags := make([]string, 0)
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
