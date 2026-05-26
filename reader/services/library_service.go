package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chooban/progger/reader/models"
	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
)

type LibraryService struct {
	db      *sqlx.DB
	tsidGen *TSIDGenerator
}

func NewLibraryService(db *DB, tsidGen *TSIDGenerator) *LibraryService {
	if tsidGen == nil {
		panic("TSIDGenerator is required")
	}
	return &LibraryService{db: db.DB, tsidGen: tsidGen}
}

func (s *LibraryService) EnsureLibrary(ctx context.Context, name string) (*models.Library, error) {
	var lib models.Library
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info("ensuring library", "name", name)
	err := s.db.GetContext(ctx, &lib, "SELECT * FROM libraries WHERE name = ?", name)
	if err == nil {
		return &lib, nil
	}

	now := FormatTime(time.Now().UTC())
	lib = models.Library{
		ID:        HashEntityID("lib", name),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO libraries (id, name, created_at, updated_at) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		lib.ID, lib.Name, lib.CreatedAt, lib.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create library: %w", err)
	}

	return &lib, nil
}

func (s *LibraryService) List(ctx context.Context) ([]models.Library, error) {
	var libs []models.Library
	err := s.db.SelectContext(ctx, &libs, "SELECT * FROM libraries ORDER BY name")
	if err != nil {
		return nil, err
	}
	return libs, nil
}

func (s *LibraryService) GetByID(ctx context.Context, id int64) (*models.Library, error) {
	var lib models.Library
	err := s.db.GetContext(ctx, &lib, "SELECT * FROM libraries WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &lib, nil
}
