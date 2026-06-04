package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type LibraryRepo struct {
	db *sqlx.DB
}

func NewLibraryRepo(db *DB) *LibraryRepo {
	return &LibraryRepo{db: db.DB}
}

func (r *LibraryRepo) EnsureLibrary(ctx context.Context, name string) (*Library, error) {
	var lib Library
	err := r.db.GetContext(ctx, &lib, "SELECT * FROM libraries WHERE name = ?", name)
	if err == nil {
		return &lib, nil
	}

	now := formatTime(time.Now().UTC())
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO libraries (name, created_at, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING
	`, name, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert library: %w", err)
	}

	err = r.db.GetContext(ctx, &lib, "SELECT * FROM libraries WHERE name = ?", name)
	if err != nil {
		return nil, fmt.Errorf("failed to get inserted library: %w", err)
	}

	return &lib, nil
}

func (r *LibraryRepo) List(ctx context.Context) ([]Library, error) {
	var libs []Library
	err := r.db.SelectContext(ctx, &libs, "SELECT id, name, created_at, updated_at FROM libraries ORDER BY name")
	if err != nil {
		return nil, err
	}
	return libs, nil
}

func (r *LibraryRepo) GetByID(ctx context.Context, id int64) (*Library, error) {
	var lib Library
	err := r.db.GetContext(ctx, &lib, "SELECT id, name, created_at, updated_at FROM libraries WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &lib, nil
}
