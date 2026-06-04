package database

import (
	"context"
	_ "embed"
	"strings"
	"time"

	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	*sqlx.DB
}

var defaultKnownTitles = []string{
	"Anderson, Psi-Division",
	"Chimpsky's Law",
	"Counterfeit Girl",
	"Feral & Foe",
	"Lowborn High",
	"Scarlet Traces",
	"Strontium Dog",
	"Strontium Dug",
	"The Fall of Deadworld",
}

var defaultSkipTitles = []string{
	"Interrogation",
	"New Books",
	"Obituary",
	"Tribute",
	"Untitled",
	"Encyclopedia",
	"Prog Finished",
	"A Year in Thrills",
}

func Open(path string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", path+"?_fk=1")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	d := &DB{DB: db}
	if err := d.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := d.seedDefaults(); err != nil {
		db.Close()
		return nil, err
	}

	return d, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

func (d *DB) migrate() error {
	ctx := context.Background()
	statements := strings.Split(schemaSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := d.ExecContext(ctx, stmt)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (d *DB) seedDefaults() error {
	ctx := context.Background()
	var count int
	if err := d.GetContext(ctx, &count, "SELECT COUNT(*) FROM known_titles"); err != nil {
		return err
	}
	if count == 0 {
		for _, title := range defaultKnownTitles {
			_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO known_titles (name) VALUES (?)", title)
			if err != nil {
				return err
			}
		}
	}
	if err := d.GetContext(ctx, &count, "SELECT COUNT(*) FROM skip_titles"); err != nil {
		return err
	}
	if count == 0 {
		for _, title := range defaultSkipTitles {
			_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO skip_titles (name) VALUES (?)", title)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}
