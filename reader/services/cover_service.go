package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/chooban/progger/reader/models"
	"github.com/chooban/progger/scan/api"
	"github.com/jmoiron/sqlx"
)

type CoverService struct {
	db      *sqlx.DB
	tsidGen *TSIDGenerator
}

func NewCoverService(db *DB, tsidGen *TSIDGenerator) *CoverService {
	if tsidGen == nil {
		panic("TSIDGenerator is required")
	}
	return &CoverService{db: db.DB, tsidGen: tsidGen}
}

func (s *CoverService) Upsert(ctx context.Context, cover *models.Cover) error {
	now := FormatTime(time.Now().UTC())
	cover.CreatedAt = now

	if cover.ID == 0 {
		tsid, err := s.tsidGen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate cover ID: %w", err)
		}
		cover.ID = tsid
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO covers (id, text, series_id, artist, filename, publication, issue_number, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`,
		cover.ID, cover.Text, cover.SeriesID, cover.Artist, cover.Filename, cover.Publication, cover.IssueNumber, cover.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert cover: %w", err)
	}

	return nil
}

func (s *CoverService) UpsertFromAPI(ctx context.Context, cover api.Cover, issue api.Issue, series *models.Series) (*models.Cover, error) {
	c := &models.Cover{
		ID:          HashEntityID("cover", series.Name, strconv.Itoa(issue.IssueNumber)),
		Text:        cover.Text,
		SeriesID:    series.ID,
		Artist:      cover.Artist,
		Filename:    cover.Filename,
		IssueNumber: issue.IssueNumber,
		Publication: issue.Publication,
	}
	if err := s.Upsert(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CoverService) GetByID(ctx context.Context, id int64) (*models.Cover, error) {
	var cover models.Cover
	err := s.db.GetContext(ctx, &cover, "SELECT * FROM covers WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (s *CoverService) ListAll(ctx context.Context, offset, limit int) ([]models.Cover, int, error) {
	var covers []models.Cover
	var total int

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM covers")
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM covers ORDER BY created_at DESC", nil, limit, offset)

	err := s.db.SelectContext(ctx, &covers, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return covers, total, nil
}

func (s *CoverService) FindBySeriesName(ctx context.Context, seriesName string) (*models.Cover, error) {
	seriesSer := &SeriesService{db: s.db}
	series, err := seriesSer.MaybeGetByName(ctx, seriesName)
	if err != nil {
		return nil, err
	}
	covers, err := s.FindAllBySeries(ctx, series.ID)
	if err != nil {
		return nil, err
	}
	if len(covers) == 0 {
		return nil, sql.ErrNoRows
	}
	return &covers[0], nil
}

func (s *CoverService) FindAllBySeries(ctx context.Context, seriesID int64) ([]models.Cover, error) {
	var covers []models.Cover
	err := s.db.SelectContext(ctx, &covers, "SELECT * FROM covers WHERE series_id = ? ORDER BY issue_number", seriesID)
	if err != nil {
		return nil, err
	}
	return covers, nil
}

func (s *CoverService) FindFirstBySeries(ctx context.Context, seriesID int64) (*models.Cover, error) {
	var cover models.Cover
	err := s.db.GetContext(ctx, &cover, "SELECT * FROM covers WHERE series_id = ? ORDER BY issue_number ASC LIMIT 1", seriesID)
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

// FindForBook will return a cover for a book if one exists. It'll look for any cover for the
// book's series that appears during the books run.
func (s *CoverService) FindForBook(ctx context.Context, book *models.Book) ([]*models.Cover, error) {
	var covers []*models.Cover
	query := `
select c.*
from books b
join episodes e on (e.book_id = b.id)
join covers c on (e.issue_number = c.issue_number and c.series_id = b.series_id)
where b.id = ?
ORDER by e.issue_number ASC, e.part ASC;
`

	err := s.db.SelectContext(ctx, &covers, query, book.ID)

	return covers, err
}
