package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type SeriesRepo struct {
	db *sqlx.DB
}

func NewSeriesRepo(db *DB) *SeriesRepo {
	return &SeriesRepo{db: db.DB}
}

func (r *SeriesRepo) List(ctx context.Context, offset, limit int) ([]Series, int, error) {
	var series []Series
	var total int

	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series"); err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM series ORDER BY name", nil, limit, offset)
	err := r.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (r *SeriesRepo) ListByLibraryID(ctx context.Context, libraryID int64, offset, limit int) ([]Series, int, error) {
	var series []Series
	var total int

	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series WHERE library_id = ?", libraryID); err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM series WHERE library_id = ? ORDER BY name", []any{libraryID}, limit, offset)
	err := r.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (r *SeriesRepo) GetByID(ctx context.Context, id int64) (*Series, error) {
	var series Series
	err := r.db.GetContext(ctx, &series, "SELECT * FROM series WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &series, nil
}

func (r *SeriesRepo) MaybeGetByName(ctx context.Context, name string) (*Series, error) {
	var series Series
	err := r.db.GetContext(ctx, &series, "SELECT * FROM series WHERE name = ?", name)
	if err != nil {
		return nil, err
	}
	return &series, nil
}

func (r *SeriesRepo) Upsert(ctx context.Context, series *Series) error {
	if series.ID == 0 {
		return fmt.Errorf("series ID must be set before upsert")
	}

	now := formatTime(time.Now().UTC())
	if series.CreatedAt == "" {
		series.CreatedAt = now
	}
	series.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO series (id, library_id, name, book_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			library_id = excluded.library_id,
			book_count = excluded.book_count,
			updated_at = excluded.updated_at
	`, series.ID, series.LibraryID, series.Name, series.BookCount, series.CreatedAt, series.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert series: %w", err)
	}

	return nil
}

func (r *SeriesRepo) SetBookCount(ctx context.Context, seriesID int64, count int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE series SET book_count = ? WHERE id = ?", count, seriesID)
	return err
}

func (r *SeriesRepo) ListRecentlyAdded(ctx context.Context, offset, limit int) ([]Series, int, error) {
	var series []Series
	var total int

	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series WHERE created_at >= ?", formatTime(time.Now().UTC().AddDate(0, 0, -14)))
	if err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM series WHERE created_at >= ? ORDER BY created_at DESC",
		[]any{formatTime(time.Now().UTC().AddDate(0, 0, -14))}, limit, offset)
	err = r.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (r *SeriesRepo) ListRecentlyUpdated(ctx context.Context, offset, limit int) ([]Series, int, error) {
	var series []Series
	var total int

	cutoff := formatTime(time.Now().UTC().AddDate(0, 0, -14))
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series WHERE updated_at >= ? OR created_at >= ?", cutoff, cutoff)
	if err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM series WHERE updated_at >= ? OR created_at >= ? ORDER BY updated_at DESC",
		[]any{cutoff, cutoff}, limit, offset)
	err = r.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (r *SeriesRepo) ListLatest(ctx context.Context, libraryIDs []int64, offset, limit int) ([]Series, int, error) {
	var series []Series
	var total int

	var countQuery, dataQuery string
	var args []any

	if len(libraryIDs) > 0 {
		placeholders := make([]string, len(libraryIDs))
		for i := range libraryIDs {
			placeholders[i] = "?"
		}
		inClause := "(" + strings.Join(placeholders, ",") + ")"
		countQuery = "SELECT COUNT(*) FROM series WHERE library_id IN " + inClause
		dataQuery = "SELECT * FROM series WHERE library_id IN " + inClause + " ORDER BY created_at DESC"
		for _, id := range libraryIDs {
			args = append(args, id)
		}
	} else {
		countQuery = "SELECT COUNT(*) FROM series"
		dataQuery = "SELECT * FROM series ORDER BY created_at DESC"
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	dataQuery, args = applyPagination(dataQuery, args, limit, offset)
	err := r.db.SelectContext(ctx, &series, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func applyPagination(query string, args []any, limit, offset int) (string, []any) {
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}
	return query, args
}
