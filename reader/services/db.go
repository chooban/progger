package services

import (
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var schemaSQL string

//go:embed schema.sql
var schemaFile string

func init() {
	schemaSQL = schemaFile
}

type DB struct {
	*sqlx.DB
}

func OpenDatabase(ctx context.Context, path string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", path+"?_fk=1")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return &DB{DB: db}, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

func (d *DB) Migrate(ctx context.Context) error {
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

func FormatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func ParseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}
