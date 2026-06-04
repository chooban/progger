package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type KnownTitlesRepo struct {
	db *sqlx.DB
}

func NewKnownTitlesRepo(db *DB) *KnownTitlesRepo {
	return &KnownTitlesRepo{db: db.DB}
}

func (r *KnownTitlesRepo) List(ctx context.Context) ([]string, error) {
	var titles []string
	err := r.db.SelectContext(ctx, &titles, "SELECT name FROM known_titles ORDER BY name")
	if err != nil {
		return nil, err
	}
	return titles, nil
}

func (r *KnownTitlesRepo) Add(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO known_titles (name) VALUES (?)", name)
	return err
}

func (r *KnownTitlesRepo) Delete(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM known_titles WHERE name = ?", name)
	return err
}
