package services

import (
	"context"
	"testing"
	"time"

	"github.com/chooban/progger/reader/models"
	"github.com/jmoiron/sqlx"
)

func createTestTSIDGenerator(t *testing.T) *TSIDGenerator {
	t.Helper()
	tsidGen, err := GetTSIDGenerator()
	if err != nil {
		t.Fatalf("failed to create TSID generator for tests: %v", err)
	}
	return tsidGen
}

func CreateTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	_, err = db.Exec(schemaSQL)
	if err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

func CreateTestLibrary(t *testing.T, service *LibraryService, name string) *models.Library {
	t.Helper()

	// Generate TSID for test library
	tsid, err := service.tsidGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate library TSID: %v", err)
	}

	lib := &models.Library{
		ID:        tsid,
		Name:      name,
		CreatedAt: FormatTime(time.Now().UTC()),
		UpdatedAt: FormatTime(time.Now().UTC()),
	}

	_, err = service.db.ExecContext(context.Background(),
		"INSERT INTO libraries (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		lib.ID, lib.Name, lib.CreatedAt, lib.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to create test library: %v", err)
	}

	return lib
}

func CreateTestSeries(t *testing.T, service *SeriesService, libraryID int64) *models.Series {
	return CreateTestSeriesWithName(t, service, libraryID, "Test Series")
}

func CreateTestCover(t *testing.T, service *CoverService, seriesId int64, issueNumber int) *models.Cover {
	t.Helper()

	// Generate TSID for test cover
	tsid, err := service.tsidGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate cover TSID: %v", err)
	}

	cover := &models.Cover{
		ID:          tsid,
		IssueNumber: issueNumber,
		SeriesID:    seriesId,
		CreatedAt:   FormatTime(time.Now()),
	}

	err = service.Upsert(context.Background(), cover)
	if err != nil {
		t.Fatalf("failed to create test cover: %v", err)
	}

	return cover
}

func CreateTestSeriesWithName(t *testing.T, service *SeriesService, libraryID int64, name string) *models.Series {
	t.Helper()

	// Generate TSID for test series
	tsid, err := service.tsidGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate series TSID: %v", err)
	}

	series := &models.Series{
		ID:        tsid,
		LibraryID: libraryID,
		Name:      name,
		BookCount: 0,
		CreatedAt: FormatTime(time.Now()),
		UpdatedAt: FormatTime(time.Now()),
	}

	_, err = service.db.ExecContext(context.Background(),
		"INSERT INTO series (id, library_id, name, book_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		series.ID, series.LibraryID, series.Name, series.BookCount, series.CreatedAt, series.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to create test series: %v", err)
	}

	return series
}

func CreateTestBook(t *testing.T, service *BookService, name string, seriesID int64) *models.Book {
	t.Helper()

	// Generate TSID for test book
	tsid, err := service.tsidGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate book TSID: %v", err)
	}

	book := &models.Book{
		ID:          tsid,
		SeriesID:    seriesID,
		Name:        name,
		Number:      1,
		Publication: "Test Publication",
		Status:      "READY",
		FileHash:    "testhash",
		PageCount:   100,
		FirstIssue:  1,
		LastIssue:   1,
		CreatedAt:   FormatTime(time.Now()),
		UpdatedAt:   FormatTime(time.Now()),
	}

	_, err = service.db.ExecContext(context.Background(),
		`INSERT INTO books (id, series_id, name, number, publication, file_hash, status, page_count, first_issue, last_issue, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		book.ID, book.SeriesID, book.Name, book.Number, book.Publication, book.FileHash,
		book.Status, book.PageCount, book.FirstIssue, book.LastIssue, book.CreatedAt, book.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}

	return book
}

type testFixture struct {
	DB             *sqlx.DB
	DBService      *DB
	TSIDGen        *TSIDGenerator
	LibraryService *LibraryService
	SeriesService  *SeriesService
	BookService    *BookService
	CoverService   *CoverService
}

func setupTestFixture(t *testing.T) *testFixture {
	t.Helper()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)

	return &testFixture{
		DB:             db,
		DBService:      dbService,
		TSIDGen:        tsidGen,
		LibraryService: NewLibraryService(dbService, tsidGen),
		SeriesService:  NewSeriesService(dbService, tsidGen),
		BookService:    NewBookService(dbService, tsidGen),
		CoverService:   NewCoverService(dbService, tsidGen),
	}
}

func CreateTestEpisode(t *testing.T, service *BookService, bookID int64, issueNumber, part int) *models.Episode {
	t.Helper()

	// Generate TSID for test episode
	tsid, err := service.tsidGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate episode TSID: %v", err)
	}

	episode := &models.Episode{
		ID:          tsid,
		BookID:      bookID,
		Filename:    "test.pdf",
		IssueNumber: issueNumber,
		Title:       "Test Episode",
		Part:        part,
		PageFrom:    1,
		PageTo:      10,
	}

	_, err = service.db.ExecContext(context.Background(),
		`INSERT INTO episodes (id, book_id, filename, issue_number, title, part, page_from, page_to)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		episode.ID, episode.BookID, episode.Filename, episode.IssueNumber, episode.Title, episode.Part,
		episode.PageFrom, episode.PageTo)
	if err != nil {
		t.Fatalf("failed to create test episode: %v", err)
	}

	return episode
}
