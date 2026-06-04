package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SkipTitlesRepo struct {
	db *sqlx.DB
}

func NewSkipTitlesRepo(db *DB) *SkipTitlesRepo {
	return &SkipTitlesRepo{db: db.DB}
}

func (r *SkipTitlesRepo) List(ctx context.Context) ([]string, error) {
	var titles []string
	err := r.db.SelectContext(ctx, &titles, "SELECT name FROM skip_titles ORDER BY name")
	if err != nil {
		return nil, err
	}
	return titles, nil
}

func (r *SkipTitlesRepo) Add(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO skip_titles (name) VALUES (?)", name)
	return err
}

func (r *SkipTitlesRepo) Delete(ctx context.Context, name string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM skip_titles WHERE name = ?", name)
	return err
}
