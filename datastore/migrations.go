package datastore

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type migration struct {
	date   string
	number int
}

func (m1 migration) before(m2 migration) bool {
	return m1.date < m2.date || (m1.date == m2.date && m1.number < m2.number)
}

func (ds *Datastore) RunMigrations(migrations fs.FS) (uint, error) {
	// initialize _migration table
	_, err := ds.db.Exec(`create table if not exists _migration (
		date	text,
		number	number,
		primary key (date, number))`)
	if err != nil {
		return 0, fmt.Errorf("creating _migration: %w", err)
	}

	var migrationList []migration
	rows, err := ds.db.Query(`select * from _migration`)
	for rows.Next() {
		var m migration
		rows.Scan(&m.date, &m.number)
		migrationList = append(migrationList, m)
	}
	if err != nil {
		return 0, fmt.Errorf("reading in migrations: %w", err)
	}

	// stores list of migrations & whether they've been applied
	migrationSet := make(map[migration]bool)
	latestMigration := migration{"", 0}

	if len(migrationList) == 0 {
		log.Println("Creating database")
	} else {
		for _, m := range migrationList {
			migrationSet[m] = false
			if latestMigration.before(m) {
				latestMigration = m
			}
		}
	}

	var migrationsPerformed uint = 0

	// Execute only migrations which are more recent than the latest migration
	// Because walkdir traverses the directory in lexicographical order, we don't need
	// to worry about the migrations being performed in the wrong order
	err = fs.WalkDir(migrations, ".", func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking dir: %s", err)
		}
		if d.IsDir() {
			return nil
		}

		// parse & validate migration name
		filename := path.Base(filepath)
		newMigration, err := parseMigrationName(filename)
		if err != nil {
			return fmt.Errorf("parsing migration name: %w", err)
		}

		// if migration is already in db
		isRun, exists := migrationSet[newMigration]
		if exists {
			if isRun {
				return fmt.Errorf("duplicate migration: %s.%d has already been visited", newMigration.date, newMigration.number)
			}
			if latestMigration.before(newMigration) {
				// this migration is somehow out of order
				return fmt.Errorf("migrations are out of order: old migration %s.%d is newer than latest migration %s.%d",
					newMigration.date, newMigration.number, latestMigration.date, latestMigration.number)
			}
			// skip old migration
			migrationSet[newMigration] = true
			return nil

		} else {
			// this is a new migration
			if !latestMigration.before(newMigration) {
				return fmt.Errorf("migrations are out of order: new migration %s.%d is no newer than than latest migration %s.%d",
					newMigration.date, newMigration.number, latestMigration.date, latestMigration.number)
			}
			// this is our new latest migration; continue as normal
			migrationSet[newMigration] = true
			latestMigration = newMigration
		}

		// do migration
		file, err := fs.ReadFile(migrations, filepath)
		if err != nil {
			return fmt.Errorf("reading migration file %s: %s", filepath, err)
		}
		err = ds.runMigration(newMigration, string(file))
		if err != nil {
			return fmt.Errorf("running migration %s.%d: %s", newMigration.date, newMigration.number, err)
		}
		migrationsPerformed += 1

		return nil
	})

	for m, done := range migrationSet {
		if !done {
			return migrationsPerformed, fmt.Errorf("missing migration: migration %s.%d was never visited", m.date, m.number)
		}
	}
	return migrationsPerformed, err
}

func parseMigrationName(filename string) (migration, error) {
	match, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2}\.\d+\.sql$`, filename)
	if err != nil {
		return migration{}, fmt.Errorf("validating name: %s", err)
	}
	if !match {
		return migration{}, fmt.Errorf("bad format %s", filename)
	}
	nameComponents := strings.Split(filename, ".")
	if len(nameComponents) != 3 {
		return migration{}, fmt.Errorf("bad format %s", filename)
	}
	date := nameComponents[0]
	number64, err := strconv.ParseInt(nameComponents[1], 10, 32)
	if err != nil {
		return migration{}, fmt.Errorf("number too large %s", filename)
	}
	number := int(number64)
	return migration{date, number}, nil
}

func (ds *Datastore) runMigration(name migration, contents string) error {
	ctx, stop := context.WithCancel(context.Background())
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		stop()
		return fmt.Errorf("beginning transaction: %s", err)
	}
	_, err = tx.Exec(contents)
	if err != nil {
		stop()
		return fmt.Errorf("executing migration: %s", err)
	}
	_, err = tx.Exec(`insert into _migration values (?, ?)`, name.date, name.number)
	if err != nil {
		stop()
		return fmt.Errorf("inserting migration into log table: %s", err)
	}
	err = tx.Commit()
	stop()
	if err != nil {
		return fmt.Errorf("committing migration: %s", err)
	}
	return nil
}
