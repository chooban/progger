package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type BookRepo struct {
	db *sqlx.DB
}

func NewBookRepo(db *DB) *BookRepo {
	return &BookRepo{db: db.DB}
}

func (r *BookRepo) addEpisodesToBooks(ctx context.Context, books []*Book) error {
	if len(books) == 0 {
		return nil
	}

	var booksMap = make(map[int64]*Book)
	bookIDs := make([]int64, len(books))
	for i, b := range books {
		booksMap[b.ID] = b
		bookIDs[i] = b.ID
		b.Episodes = make([]*Episode, 0)
	}

	var episodes []Episode
	query, args, err := sqlx.In("SELECT * FROM episodes WHERE book_id IN (?) ORDER BY part ASC", bookIDs)
	if err != nil {
		return fmt.Errorf("failed to expand IN query: %w", err)
	}
	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &episodes, query, args...)
	if err != nil {
		return fmt.Errorf("failed to select episodes: %w", err)
	}

	for i := range episodes {
		ep := episodes[i]
		if book, ok := booksMap[ep.BookID]; ok {
			book.Episodes = append(book.Episodes, &ep)
		}
	}

	return nil
}

func (r *BookRepo) CountBySeries(ctx context.Context, seriesID int64) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM books WHERE series_id = ?", seriesID)
	return count, err
}

func (r *BookRepo) ListBySeries(ctx context.Context, seriesID int64, offset, limit int) ([]*Book, int, error) {
	var books []*Book
	var total int

	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books WHERE series_id = ?", seriesID); err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM books WHERE series_id = ? ORDER BY number ASC", []any{seriesID}, limit, offset)
	err := r.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepo) ListAll(ctx context.Context, offset, limit int) ([]*Book, int, error) {
	var books []*Book
	var total int

	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books"); err != nil {
		return nil, 0, err
	}

	query, args := applyPagination("SELECT * FROM books ORDER BY number ASC", nil, limit, offset)
	err := r.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepo) ListByLibraryID(ctx context.Context, libraryID int64, offset, limit int) ([]*Book, int, error) {
	var books []*Book
	var total int

	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM books b
		JOIN series s ON b.series_id = s.id
		WHERE s.library_id = ?
	`, libraryID); err != nil {
		return nil, 0, err
	}

	query, args := applyPagination(`
		SELECT b.* FROM books b
		JOIN series s ON b.series_id = s.id
		WHERE s.library_id = ?
		ORDER BY b.number ASC
	`, []any{libraryID}, limit, offset)
	err := r.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepo) ListOnDeck(ctx context.Context, libraryIDs []int64, offset, limit int) ([]*Book, int, error) {
	var books []*Book
	var total int

	var args []any
	var whereClause string

	if len(libraryIDs) > 0 {
		placeholders := make([]string, len(libraryIDs))
		for i, id := range libraryIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClause = "WHERE s.library_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM books b
		JOIN series s ON b.series_id = s.id
		%s
		AND b.id = (
			SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
		)
	`, whereClause)

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT b.* FROM books b
		JOIN series s ON b.series_id = s.id
		%s
		AND b.id = (
			SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
		)
		ORDER BY b.name ASC
	`, whereClause)

	dataQuery, args = applyPagination(dataQuery, args, limit, offset)
	err := r.db.SelectContext(ctx, &books, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepo) GetByID(ctx context.Context, id int64) (*Book, error) {
	var book Book
	err := r.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	episodes, err := r.getEpisodes(ctx, id)
	if err != nil {
		return nil, err
	}
	book.Episodes = episodes

	return &book, nil
}

func (r *BookRepo) getEpisodes(ctx context.Context, bookID int64) ([]*Episode, error) {
	episodes := make([]*Episode, 0)
	err := r.db.SelectContext(ctx, &episodes, "SELECT * FROM episodes WHERE book_id = ? ORDER BY part ASC", bookID)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func (r *BookRepo) Upsert(ctx context.Context, book *Book) error {
	if book.ID == 0 {
		return fmt.Errorf("book ID must be set before upsert")
	}

	now := formatTime(time.Now().UTC())
	if book.CreatedAt == "" {
		book.CreatedAt = now
	}
	book.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO books (id, series_id, name, number, publication, size_bytes, file_hash, status, page_count, first_issue, last_issue, release_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`,
		book.ID, book.SeriesID, book.Name, book.Number, book.Publication, book.SizeBytes, book.FileHash, book.Status, book.PageCount, book.FirstIssue, book.LastIssue, book.ReleaseDate, book.CreatedAt, book.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert book: %w", err)
	}

	return nil
}

func (r *BookRepo) UpsertEpisodes(ctx context.Context, episodes []*Episode) error {
	for _, episode := range episodes {
		if episode.ID == 0 {
			return fmt.Errorf("episode ID must be set before upsert")
		}

		_, err := r.db.ExecContext(ctx, `
			INSERT OR IGNORE INTO episodes (id, book_id, filename, issue_number, title, part, page_from, page_to, release_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			episode.ID, episode.BookID, episode.Filename, episode.IssueNumber, episode.Title, episode.Part, episode.PageFrom, episode.PageTo, episode.ReleaseDate)

		if err != nil {
			return fmt.Errorf("failed to insert episode: %w", err)
		}
	}

	return nil
}

func (r *BookRepo) GetForThumbnail(ctx context.Context, bookId int64) (*BookWithSeries, error) {
	var result struct {
		Book
		Filename string `db:"filename"`
	}

	query := `SELECT b.*, e.filename as filename
		FROM books b
		LEFT JOIN episodes e ON e.book_id = b.id
		WHERE b.id = ?
		ORDER BY e.issue_number ASC, e.part ASC LIMIT 1`

	err := r.db.GetContext(ctx, &result, query, bookId)
	if err != nil {
		return nil, err
	}

	return &BookWithSeries{
		Book:     &result.Book,
		Filename: result.Filename,
	}, nil
}

func (r *BookRepo) MaybeGetBySeriesAndName(ctx context.Context, id int64, name string) (*Book, error) {
	var book Book
	err := r.db.GetContext(ctx, &book, "SELECT * FROM books WHERE series_id = ? AND name = ?", id, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &book, nil
}

func (r *BookRepo) GetEpisodeByBookAndPart(ctx context.Context, bookID int64, part int) (*Episode, error) {
	var episode Episode
	err := r.db.GetContext(ctx, &episode, "SELECT * FROM episodes WHERE book_id = ? AND part = ?", bookID, part)
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

func (r *BookRepo) ListLatest(ctx context.Context, libraryIDs []int64, offset, limit int) ([]*Book, int, error) {
	var books []*Book
	var total int

	var countQuery, dataQuery string
	var args []any

	if len(libraryIDs) > 0 {
		placeholders := make([]string, len(libraryIDs))
		for i := range libraryIDs {
			placeholders[i] = "?"
		}
		inClause := "(" + strings.Join(placeholders, ",") + ")"
		countQuery = `SELECT COUNT(*) FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id IN ` + inClause
		dataQuery = `SELECT b.* FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id IN ` + inClause + ` ORDER BY b.created_at DESC`
		for _, id := range libraryIDs {
			args = append(args, id)
		}
	} else {
		countQuery = "SELECT COUNT(*) FROM books"
		dataQuery = "SELECT * FROM books ORDER BY created_at DESC"
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	dataQuery, args = applyPagination(dataQuery, args, limit, offset)
	err := r.db.SelectContext(ctx, &books, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepo) GetNextBook(ctx context.Context, bookID int64) (*Book, error) {
	var book Book
	err := r.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", bookID)
	if err != nil {
		return nil, err
	}

	var next Book
	err = r.db.GetContext(ctx, &next, "SELECT * FROM books WHERE series_id = ? AND number > ? ORDER BY number ASC LIMIT 1", book.SeriesID, book.Number)
	if err != nil {
		return nil, err
	}

	books := []*Book{&next}
	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, err
	}
	return books[0], nil
}

func (r *BookRepo) GetPreviousBook(ctx context.Context, bookID int64) (*Book, error) {
	var book Book
	err := r.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", bookID)
	if err != nil {
		return nil, err
	}

	var prev Book
	err = r.db.GetContext(ctx, &prev, "SELECT * FROM books WHERE series_id = ? AND number < ? ORDER BY number DESC LIMIT 1", book.SeriesID, book.Number)
	if err != nil {
		return nil, err
	}

	books := []*Book{&prev}
	if err := r.addEpisodesToBooks(ctx, books); err != nil {
		return nil, err
	}
	return books[0], nil
}
