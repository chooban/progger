package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type CoverRepo struct {
	db *sqlx.DB
}

func NewCoverRepo(db *DB) *CoverRepo {
	return &CoverRepo{db: db.DB}
}

func (r *CoverRepo) Upsert(ctx context.Context, cover *Cover) error {
	if cover.ID == 0 {
		return fmt.Errorf("cover ID must be set before upsert")
	}

	cover.CreatedAt = formatTime(time.Now().UTC())

	_, err := r.db.ExecContext(ctx, `
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

func (r *CoverRepo) GetByID(ctx context.Context, id int64) (*Cover, error) {
	var cover Cover
	err := r.db.GetContext(ctx, &cover, "SELECT * FROM covers WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (r *CoverRepo) ListAll(ctx context.Context, offset, limit int) ([]Cover, int, error) {
	var covers []Cover
	var total int

	countErr := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM covers")
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM covers ORDER BY created_at DESC", nil, limit, offset)
	err := r.db.SelectContext(ctx, &covers, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return covers, total, nil
}

func (r *CoverRepo) FindBySeriesName(ctx context.Context, seriesName string) (*Cover, error) {
	var series Series
	err := r.db.GetContext(ctx, &series, "SELECT * FROM series WHERE name = ?", seriesName)
	if err != nil {
		return nil, err
	}

	covers, err := r.FindAllBySeries(ctx, series.ID)
	if err != nil {
		return nil, err
	}
	if len(covers) == 0 {
		return nil, fmt.Errorf("no covers found for series: %s", seriesName)
	}
	return &covers[0], nil
}

func (r *CoverRepo) FindAllBySeries(ctx context.Context, seriesID int64) ([]Cover, error) {
	var covers []Cover
	err := r.db.SelectContext(ctx, &covers, "SELECT * FROM covers WHERE series_id = ? ORDER BY issue_number", seriesID)
	if err != nil {
		return nil, err
	}
	return covers, nil
}

func (r *CoverRepo) FindFirstBySeries(ctx context.Context, seriesID int64) (*Cover, error) {
	var cover Cover
	err := r.db.GetContext(ctx, &cover, "SELECT * FROM covers WHERE series_id = ? ORDER BY issue_number ASC LIMIT 1", seriesID)
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (r *CoverRepo) FindForBook(ctx context.Context, book *Book) ([]*Cover, error) {
	var covers []*Cover
	query := `
select c.*
from books b
join episodes e on (e.book_id = b.id)
join covers c on (e.issue_number = c.issue_number and c.series_id = b.series_id)
where b.id = ?
ORDER by e.issue_number ASC, e.part ASC;
`

	err := r.db.SelectContext(ctx, &covers, query, book.ID)
	return covers, err
}
