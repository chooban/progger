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

type BookService struct {
	db      *sqlx.DB
	tsidGen *TSIDGenerator
}

func NewBookService(db *DB, tsidGen *TSIDGenerator) *BookService {
	if tsidGen == nil {
		panic("TSIDGenerator is required")
	}
	return &BookService{db: db.DB, tsidGen: tsidGen}
}

func (s *BookService) addEpisodesToBooks(ctx context.Context, books []*models.Book) error {
	if len(books) == 0 {
		return nil
	}
	var booksMap = make(map[int64]*models.Book)
	for _, book := range books {
		booksMap[book.ID] = book
	}
	var bookIds = make([]int64, len(books))
	for i, book := range books {
		bookIds[i] = book.ID
	}
	var episodes []models.Episode
	query, args, err := sqlx.In("SELECT * FROM episodes WHERE book_id IN (?) ORDER by book_id, part", bookIds)
	if err != nil {
		return err
	}
	query = s.db.Rebind(query)
	err = s.db.SelectContext(ctx, &episodes, query, args...)

	for _, episode := range episodes {
		if book, ok := booksMap[episode.BookID]; ok {
			book.Episodes = append(book.Episodes, &episode)
		}
	}

	return err
}

func (s *BookService) CountBySeries(ctx context.Context, seriesID int64) (int, error) {
	var total int
	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books WHERE series_id = ?", seriesID)
	if countErr != nil {
		return 0, countErr
	}
	return total, nil

}
func (s *BookService) ListBySeries(ctx context.Context, seriesID int64, offset, limit int) ([]*models.Book, int, error) {
	var books []*models.Book
	var total int

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books WHERE series_id = ?", seriesID)
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM books WHERE series_id = ? ORDER BY number ASC", []any{seriesID}, limit, offset)

	err := s.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) ListAll(ctx context.Context, offset, limit int) ([]*models.Book, int, error) {
	var books []*models.Book
	var total int

	countErr := s.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM books")
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination("SELECT * FROM books ORDER BY number ASC", nil, limit, offset)

	err := s.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	err = s.addEpisodesToBooks(ctx, books)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) ListByLibraryID(ctx context.Context, libraryID int64, offset, limit int) ([]*models.Book, int, error) {
	var books []*models.Book
	var total int

	countErr := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id = ?`, libraryID)
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args := applyPagination(`SELECT b.* FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id = ? ORDER BY b.number ASC`, []any{libraryID}, limit, offset)

	err := s.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	err = s.addEpisodesToBooks(ctx, books)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) ListOnDeck(ctx context.Context, libraryIDs []int64, offset, limit int) ([]*models.Book, int, error) {
	var books []*models.Book
	var total int

	var query string
	var countQuery string
	var args []any

	if len(libraryIDs) == 0 {
		countQuery = `SELECT COUNT(*) FROM books b WHERE b.id = (
			SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
		)`
		query = `SELECT b.* FROM books b WHERE b.id = (
			SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
		) ORDER BY b.name`
	} else {
		placeholders := make([]string, len(libraryIDs))
		for i := range libraryIDs {
			placeholders[i] = "?"
		}
		libIDStr := "WHERE s.library_id IN (" + strings.Join(placeholders, ",") + ")"

		countQuery = `SELECT COUNT(*) FROM books b 
			JOIN series s ON b.series_id = s.id ` + libIDStr + `
			AND b.id = (
				SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
			)`
		query = `SELECT b.* FROM books b
			JOIN series s ON b.series_id = s.id ` + libIDStr + `
			AND b.id = (
				SELECT MIN(b2.id) FROM books b2 WHERE b2.series_id = b.series_id
			) ORDER BY b.name`

		for _, id := range libraryIDs {
			args = append(args, id)
		}
	}

	countErr := s.db.GetContext(ctx, &total, countQuery, args...)
	if countErr != nil {
		return nil, 0, countErr
	}

	query, args = applyPagination(query, args, limit, offset)

	err := s.db.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, 0, err
	}

	err = s.addEpisodesToBooks(ctx, books)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	var book models.Book
	err := s.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	episodes, err := s.getEpisodes(ctx, book.ID)
	if err != nil {
		return nil, err
	}

	book.Episodes = episodes

	return &book, nil
}

func (s *BookService) getEpisodes(ctx context.Context, bookID int64) ([]*models.Episode, error) {
	episodes := make([]*models.Episode, 0)
	err := s.db.SelectContext(ctx, &episodes,
		"SELECT * FROM episodes WHERE book_id = ? ORDER BY issue_number ASC, part ASC", bookID)
	return episodes, err
}

func (s *BookService) Upsert(ctx context.Context, book *models.Book) error {
	logger := logr.FromContextOrDiscard(ctx)

	now := FormatTime(time.Now().UTC())
	book.UpdatedAt = now
	if book.ID == 0 {
		logger.Info("creating new book", "series_id", book.SeriesID, "name", book.Name)
		tsid, err := s.tsidGen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate book ID: %w", err)
		}
		book.ID = tsid
		book.CreatedAt = now
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO books (id, series_id, name, publication, number, size_bytes, file_hash, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`,
		book.ID, book.SeriesID, book.Name, book.Publication, book.Number, book.SizeBytes, book.FileHash,
		book.Status, book.CreatedAt, book.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert book: %w", err)
	}

	return nil
}

func (s *BookService) UpsertEpisodes(ctx context.Context, episodes []*models.Episode) error {
	logger := logr.FromContextOrDiscard(ctx)
	for _, episode := range episodes {
		if episode.BookID == 0 {
			return errors.New("episode has no bookID")
		}
		if episode.ID == 0 {
			logger.Info("looking for existing episode", "book_id", episode.BookID, "part", episode.Part)
			existing, err := s.GetEpisodeByBookAndPart(ctx, episode.BookID, episode.Part)
			if err == nil && existing != nil {
				logger.Info("found existing episode", "episode_id", existing.ID)
				episode.ID = existing.ID
			} else {
				logger.Info("Generated ID for episode")
				if tsid, err := s.tsidGen.Generate(); err != nil {
					return fmt.Errorf("failed to generate episode ID: %w", err)
				} else {
					episode.ID = tsid
				}
			}
		}

		logger.V(1).Info("attempting to insert episode", "book_id", episode.BookID, "part", episode.Part, "issue_number", episode.IssueNumber)

		_, err := s.db.ExecContext(ctx, `
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

func (s *BookService) GetForThumbnail(ctx context.Context, bookId int64) (*models.BookWithSeries, error) {
	var result struct {
		models.Book
		Filename string `db:"filename"`
	}

	query := `SELECT b.*, e.filename as filename 
		FROM books b
		LEFT JOIN episodes e ON e.book_id = b.id
		WHERE b.id = ? 
		ORDER BY e.issue_number ASC, e.part ASC LIMIT 1`

	err := s.db.GetContext(ctx, &result, query, bookId)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	return &models.BookWithSeries{
		Book:     &result.Book,
		Filename: result.Filename,
	}, nil
}

func (s *BookService) MaybeGetBySeriesAndName(ctx context.Context, id int64, name string) (*models.Book, error) {
	if id == 0 {
		return nil, errors.New("zero passed as book id")
	}
	var book models.Book
	err := s.db.GetContext(ctx, &book, "SELECT * FROM books WHERE series_id = ? AND name = ?", id, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &book, nil

}

func (s *BookService) GetEpisodeByBookAndPart(ctx context.Context, bookID int64, part int) (*models.Episode, error) {
	logger := logr.FromContextOrDiscard(ctx)
	var episode models.Episode

	logger.Info("getting episode", "book_id", bookID, "part", part)
	if err := s.db.GetContext(ctx, &episode, "SELECT * FROM episodes WHERE book_id = ? AND part = ?", bookID, part); err != nil {
		return nil, err
	} else {
		logger.Info("found episode", "book_id", bookID, "part", part, "episode_id", episode.ID)
		return &episode, nil
	}
}

func (s *BookService) ListLatest(ctx context.Context, libraryIDs []int64, offset, limit int) ([]*models.Book, int, error) {
	var books []*models.Book
	var total int

	var countQuery, dataQuery string
	var args []any

	if len(libraryIDs) > 0 {
		placeholders := make([]string, len(libraryIDs))
		for i := range libraryIDs {
			placeholders[i] = "?"
		}
		inClause := "(" + strings.Join(placeholders, ",") + ")"
		countQuery = "SELECT COUNT(*) FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id IN " + inClause
		dataQuery = "SELECT b.* FROM books b JOIN series s ON b.series_id = s.id WHERE s.library_id IN " + inClause + " ORDER BY b.created_at DESC"
		for _, id := range libraryIDs {
			args = append(args, id)
		}
	} else {
		countQuery = "SELECT COUNT(*) FROM books"
		dataQuery = "SELECT * FROM books ORDER BY created_at DESC"
	}

	countErr := s.db.GetContext(ctx, &total, countQuery, args...)
	if countErr != nil {
		return nil, 0, countErr
	}

	dataQuery, args = applyPagination(dataQuery, args, limit, offset)

	err := s.db.SelectContext(ctx, &books, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	err = s.addEpisodesToBooks(ctx, books)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) GetNextBook(ctx context.Context, bookID int64) (*models.Book, error) {
	var book models.Book
	err := s.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", bookID)
	if err != nil {
		return nil, err
	}

	var next models.Book
	err = s.db.GetContext(ctx, &next, "SELECT * FROM books WHERE series_id = ? AND number > ? ORDER BY number ASC LIMIT 1", book.SeriesID, book.Number)
	if err != nil {
		return nil, err
	}

	books := []*models.Book{&next}
	if err := s.addEpisodesToBooks(ctx, books); err != nil {
		return nil, err
	}

	return &next, nil
}

func (s *BookService) GetPreviousBook(ctx context.Context, bookID int64) (*models.Book, error) {
	var book models.Book
	err := s.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = ?", bookID)
	if err != nil {
		return nil, err
	}

	var prev models.Book
	err = s.db.GetContext(ctx, &prev, "SELECT * FROM books WHERE series_id = ? AND number < ? ORDER BY number DESC LIMIT 1", book.SeriesID, book.Number)
	if err != nil {
		return nil, err
	}

	books := []*models.Book{&prev}
	if err := s.addEpisodesToBooks(ctx, books); err != nil {
		return nil, err
	}

	return &prev, nil
}
