package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoverRepo_Upsert(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	coverRepo := NewCoverRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	cover := &Cover{
		ID:          HashEntityID("cover", "Dredd", "1"),
		SeriesID:    series.ID,
		IssueNumber: 1,
		Publication: "2000 AD",
		Text:        "Cover text",
		Artist:      "Artist",
		Filename:    "cover.jpg",
	}
	err := coverRepo.Upsert(context.Background(), cover)
	require.NoError(t, err)
	require.NotEmpty(t, cover.CreatedAt)
}

func TestCoverRepo_Upsert_FailsWithoutID(t *testing.T) {
	db := createTestDB(t)
	repo := NewCoverRepo(&DB{DB: db})

	err := repo.Upsert(context.Background(), &Cover{Text: "NoID"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ID must be set")
}

func TestCoverRepo_GetByID(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	coverRepo := NewCoverRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	cover := createTestCover(t, coverRepo, series.ID, 1)

	found, err := coverRepo.GetByID(context.Background(), cover.ID)
	require.NoError(t, err)
	require.Equal(t, cover.Text, found.Text)
}

func TestCoverRepo_FindAllBySeries(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	coverRepo := NewCoverRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	createTestCover(t, coverRepo, series.ID, 1)
	createTestCover(t, coverRepo, series.ID, 2)

	covers, err := coverRepo.FindAllBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.Len(t, covers, 2)
}

func TestCoverRepo_FindBySeriesName(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	coverRepo := NewCoverRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	cover := &Cover{
		ID:          HashEntityID("cover", "Dredd", "1"),
		SeriesID:    HashEntityID("series", "Dredd"),
		IssueNumber: 1,
		Publication: "2000 AD",
	}
	err := coverRepo.Upsert(context.Background(), cover)
	require.NoError(t, err)

	found, err := coverRepo.FindBySeriesName(context.Background(), "Dredd")
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, found)
}

func TestCoverRepo_FindForBook(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})
	bookRepo := NewBookRepo(&DB{DB: db})
	coverRepo := NewCoverRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")
	book := createTestBook(t, bookRepo, series.ID, "Judge Dredd", "2000 AD")
	date := "2025-01-01"
	ep := &Episode{ID: HashEntityID("episode", "Dredd", "1-0"), BookID: book.ID, Filename: "prog1.pdf", IssueNumber: 1, Title: "Part 1", Part: 0, PageFrom: 1, PageTo: 16, ReleaseDate: &date}
	require.NoError(t, bookRepo.UpsertEpisodes(context.Background(), []*Episode{ep}))
	createTestCover(t, coverRepo, series.ID, 1)

	covers, err := coverRepo.FindForBook(context.Background(), book)
	require.NoError(t, err)
	require.Len(t, covers, 1)
}
