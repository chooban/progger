package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
)

func createTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite3", ":memory:?_fk=1")
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { db.Close() })

	for _, stmt := range []string{
		`CREATE TABLE libraries (
			id INTEGER NOT NULL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE series (
			id INTEGER NOT NULL PRIMARY KEY,
			library_id INTEGER NOT NULL REFERENCES libraries(id),
			name TEXT NOT NULL UNIQUE,
			book_count INTEGER DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(library_id, name)
		)`,
		`CREATE TABLE books (
			id INTEGER NOT NULL PRIMARY KEY,
			series_id INTEGER NOT NULL REFERENCES series(id),
			name TEXT NOT NULL,
			number INTEGER DEFAULT 0,
			publication TEXT NOT NULL,
			size_bytes INTEGER DEFAULT 0,
			file_hash TEXT,
			status TEXT DEFAULT 'READY',
			page_count INTEGER DEFAULT 0,
			first_issue INTEGER DEFAULT 0,
			last_issue INTEGER DEFAULT 0,
			release_date TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(name, series_id, publication)
		)`,
		`CREATE TABLE episodes (
			id INTEGER NOT NULL PRIMARY KEY,
			book_id INTEGER NOT NULL REFERENCES books(id),
			filename TEXT NOT NULL,
			issue_number INTEGER DEFAULT 0,
			title TEXT NOT NULL,
			part INTEGER DEFAULT 0,
			page_from INTEGER NOT NULL,
			page_to INTEGER NOT NULL,
			release_date TEXT,
			UNIQUE(book_id, issue_number, part)
		)`,
		`CREATE TABLE covers (
			id INTEGER NOT NULL PRIMARY KEY,
			issue_number INTEGER DEFAULT 0,
			publication TEXT,
			text TEXT,
			series_id INTEGER NOT NULL REFERENCES series(id),
			artist TEXT,
			filename TEXT,
			created_at TEXT NOT NULL,
			UNIQUE(issue_number, publication, series_id)
		)`,
		`CREATE TABLE known_titles (name TEXT NOT NULL PRIMARY KEY)`,
		`CREATE TABLE skip_titles (name TEXT NOT NULL PRIMARY KEY)`,
	} {
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	}

	return db
}

func createTestLibrary(t *testing.T, r *LibraryRepo, name string) *Library {
	t.Helper()
	lib, err := r.EnsureLibrary(context.Background(), name)
	if err != nil {
		t.Fatalf("failed to create test library: %v", err)
	}
	return lib
}

func createTestSeries(t *testing.T, r *SeriesRepo, libraryID int64, name string) *Series {
	t.Helper()
	series := &Series{
		ID:        HashEntityID("series", name),
		LibraryID: libraryID,
		Name:      name,
	}
	if err := r.Upsert(context.Background(), series); err != nil {
		t.Fatalf("failed to create test series: %v", err)
	}
	return series
}

func createTestBook(t *testing.T, r *BookRepo, seriesID int64, name, publication string) *Book {
	t.Helper()
	book := &Book{
		ID:          HashEntityID("book", name, publication),
		SeriesID:    seriesID,
		Name:        name,
		Publication: publication,
		Status:      "READY",
	}
	if err := r.Upsert(context.Background(), book); err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}
	return book
}

func createTestEpisode(t *testing.T, r *BookRepo, bookID int64, issueNumber, part int) *Episode {
	t.Helper()
	title := "Test Episode"
	ep := &Episode{
		ID:          HashEntityID("episode", title, fmt.Sprintf("%d-%d", issueNumber, part)),
		BookID:      bookID,
		Filename:    "test.pdf",
		IssueNumber: issueNumber,
		Title:       title,
		Part:        part,
		PageFrom:    1,
		PageTo:      32,
	}
	if err := r.UpsertEpisodes(context.Background(), []*Episode{ep}); err != nil {
		t.Fatalf("failed to create test episode: %v", err)
	}
	return ep
}

func createTestCover(t *testing.T, r *CoverRepo, seriesID int64, issueNumber int) *Cover {
	t.Helper()
	cover := &Cover{
		ID:          HashEntityID("cover", fmt.Sprintf("%d-%d", seriesID, issueNumber)),
		SeriesID:    seriesID,
		IssueNumber: issueNumber,
		Publication: "2000 AD",
		Text:        "Test Cover",
		Artist:      "Test Artist",
		Filename:    "cover.jpg",
	}
	if err := r.Upsert(context.Background(), cover); err != nil {
		t.Fatalf("failed to create test cover: %v", err)
	}
	return cover
}
