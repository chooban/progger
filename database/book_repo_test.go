package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBookRepo_Upsert(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	book := &Book{
		ID:          HashEntityID("book", "Dredd", "2000 AD"),
		SeriesID:    series.ID,
		Name:        "Judge Dredd",
		Publication: "2000 AD",
		Status:      "READY",
	}
	err := bookRepo.Upsert(context.Background(), book)
	require.NoError(t, err)
	require.NotEmpty(t, book.CreatedAt)
}

func TestBookRepo_Upsert_FailsWithoutID(t *testing.T) {
	db := createTestDB(t)
	repo := NewBookRepo(&DB{DB: db})

	err := repo.Upsert(context.Background(), &Book{Name: "NoID"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ID must be set")
}

func TestBookRepo_GetByID(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	book := createTestBook(t, bookRepo, series.ID, "Judge Dredd", "2000 AD")

	found, err := bookRepo.GetByID(context.Background(), book.ID)
	require.NoError(t, err)
	require.Equal(t, "Judge Dredd", found.Name)
	require.NotNil(t, found.Episodes)
}

func TestBookRepo_ListBySeries(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	createTestBook(t, bookRepo, series.ID, "Book A", "2000 AD")
	createTestBook(t, bookRepo, series.ID, "Book B", "2000 AD")

	books, total, err := bookRepo.ListBySeries(context.Background(), series.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, books, 2)
}

func TestBookRepo_UpsertEpisodes(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	book := createTestBook(t, bookRepo, series.ID, "Judge Dredd", "2000 AD")

	date := "2025-01-01"
	eps := []*Episode{
		{ID: HashEntityID("episode", "Dredd", "1-0"), BookID: book.ID, Filename: "prog1.pdf", IssueNumber: 1, Title: "Part 1", Part: 0, PageFrom: 1, PageTo: 16, ReleaseDate: &date},
		{ID: HashEntityID("episode", "Dredd", "1-1"), BookID: book.ID, Filename: "prog1.pdf", IssueNumber: 1, Title: "Part 2", Part: 1, PageFrom: 17, PageTo: 32, ReleaseDate: &date},
	}
	err := bookRepo.UpsertEpisodes(context.Background(), eps)
	require.NoError(t, err)

	found, err := bookRepo.GetByID(context.Background(), book.ID)
	require.NoError(t, err)
	require.Len(t, found.Episodes, 2)
}

func TestBookRepo_GetNextBook(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	bookA := createTestBook(t, bookRepo, series.ID, "Book A", "2000 AD")
	bookB := createTestBook(t, bookRepo, series.ID, "Book B", "2000 AD")

	_, err := bookRepo.db.ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", bookA.ID)
	require.NoError(t, err)
	_, err = bookRepo.db.ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", bookB.ID)
	require.NoError(t, err)

	next, err := bookRepo.GetNextBook(context.Background(), bookA.ID)
	require.NoError(t, err)
	require.Equal(t, bookB.ID, next.ID)
}

func TestBookRepo_GetForThumbnail(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	book := createTestBook(t, bookRepo, series.ID, "Judge Dredd", "2000 AD")

	date := "2025-01-01"
	ep := &Episode{ID: HashEntityID("episode", "Dredd", "thumb"), BookID: book.ID, Filename: "prog1.pdf", IssueNumber: 1, Title: "Part 1", Part: 0, PageFrom: 1, PageTo: 16, ReleaseDate: &date}
	require.NoError(t, bookRepo.UpsertEpisodes(context.Background(), []*Episode{ep}))

	result, err := bookRepo.GetForThumbnail(context.Background(), book.ID)
	require.NoError(t, err)
	require.Equal(t, "prog1.pdf", result.Filename)
	require.Equal(t, book.Name, result.Book.Name)
}

func TestBookRepo_MaybeGetBySeriesAndName(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	createTestBook(t, bookRepo, series.ID, "Judge Dredd", "2000 AD")

	found, err := bookRepo.MaybeGetBySeriesAndName(context.Background(), series.ID, "Judge Dredd")
	require.NoError(t, err)
	require.NotNil(t, found)

	missing, err := bookRepo.MaybeGetBySeriesAndName(context.Background(), series.ID, "Nonexistent")
	require.NoError(t, err)
	require.Nil(t, missing)
}
