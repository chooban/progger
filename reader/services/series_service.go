package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chooban/progger/reader/models"
	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
)

type SeriesService struct {
	db      *sqlx.DB
	tsidGen *TSIDGenerator
}

func NewSeriesService(db *DB, tsidGen *TSIDGenerator) *SeriesService {
	if tsidGen == nil {
		panic("TSIDGenerator is required")
	}
	return &SeriesService{db: db.DB, tsidGen: tsidGen}
}

func (s *SeriesService) List(ctx context.Context, offset, limit int) ([]models.Series, int, error) {
	var series []models.Series
	var total int

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series")
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM series ORDER BY name", nil, limit, offset)

	err := s.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (s *SeriesService) GetByID(ctx context.Context, id int64) (*models.Series, error) {
	var series models.Series
	err := s.db.GetContext(ctx, &series, "SELECT * FROM series WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &series, nil
}

func (s *SeriesService) MaybeGetByName(ctx context.Context, name string) (*models.Series, error) {
	var series models.Series
	err := s.db.GetContext(ctx, &series, "SELECT * FROM series WHERE name = ?", name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &series, nil
}

func (s *SeriesService) Upsert(ctx context.Context, series *models.Series) error {
	logger := logr.FromContextOrDiscard(ctx)
	now := FormatTime(time.Now().UTC())
	series.UpdatedAt = now

	logger.Info("upserting series", "series", series.Name)
	if series.ID == 0 {
		series.ID = HashEntityID("series", series.Name)
	}

	series.CreatedAt = now

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO series (id, library_id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
		    library_id = excluded.library_id,
		    updated_at = excluded.updated_at
	`,
		series.ID, series.LibraryID, series.Name,
		series.CreatedAt, series.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert series: %w", err)
	}

	return nil
}

func (s *SeriesService) SetBookCount(ctx context.Context, seriesID int64, bookCount int) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE series SET book_count = ? WHERE id = ?", bookCount, seriesID)
	return err
}

func (s *SeriesService) ListRecentlyAdded(ctx context.Context, offset, limit int) ([]models.Series, int, error) {
	var series []models.Series
	var total int

	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	cutoffDate := FormatTime(sevenDaysAgo)

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series WHERE created_at >= ?", cutoffDate)
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM series WHERE created_at >= ? ORDER BY created_at DESC", []any{cutoffDate}, limit, offset)

	err := s.db.SelectContext(ctx, &series, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}

func (s *SeriesService) ListRecentlyUpdated(ctx context.Context, offset, limit int) ([]models.Series, int, error) {
	println("ListRecentlyUpdated")
	var series []models.Series
	var total int

	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	cutoffDate := FormatTime(sevenDaysAgo)

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM series WHERE updated_at >= ? OR created_at >= ?", cutoffDate, cutoffDate)
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM series WHERE updated_at >= ? OR created_at >= ? ORDER BY updated_at DESC", []any{cutoffDate, cutoffDate}, limit, offset)

	err := s.db.SelectContext(ctx, &series, query, args...)
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

func (s *SeriesService) ListLatest(ctx context.Context, libraryIDs []int64, offset, limit int) ([]models.Series, int, error) {
	var series []models.Series
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

	countErr := s.db.GetContext(ctx, &total, countQuery, args...)
	if countErr != nil {
		return nil, 0, countErr
	}

	dataQuery, args = applyPagination(dataQuery, args, limit, offset)

	err := s.db.SelectContext(ctx, &series, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return series, total, nil
}
