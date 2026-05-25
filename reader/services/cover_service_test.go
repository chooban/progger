package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/chooban/progger/reader/models"
	"github.com/stretchr/testify/require"
)

func TestCoverService_Upsert_CreatesNewCover(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	cover := &models.Cover{
		Text:        "Test Cover",
		SeriesID:    series.ID,
		Artist:      "Test Artist",
		Filename:    "test.jpg",
		IssueNumber: 1,
		Publication: "2023-01-01",
	}

	err := f.CoverService.Upsert(context.Background(), cover)
	require.NoError(t, err)
	require.Greater(t, cover.ID, int64(0))
}

func TestCoverService_FindForBook_ReturnsCoversForBook(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create a book
	book := CreateTestBook(t, f.BookService, "Test Book", series.ID)

	// Create an episode for the book with issue number 1
	CreateTestEpisode(t, f.BookService, book.ID, 1, 0)

	// Create a cover for the series with matching issue number
	cover := &models.Cover{
		Text:        "Test Cover",
		SeriesID:    series.ID,
		Artist:      "Test Artist",
		Filename:    "test.jpg",
		IssueNumber: 1,
		Publication: "2023-01-01",
	}

	err := f.CoverService.Upsert(context.Background(), cover)
	require.NoError(t, err)

	// Find covers for the book
	covers, err := f.CoverService.FindForBook(context.Background(), book)
	require.NoError(t, err)
	require.Len(t, covers, 1)
	require.Equal(t, cover.SeriesID, covers[0].SeriesID)
}

func TestCoverService_FindForBook_ReturnsEmptyWhenNoCover(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create a book without any covers
	book := CreateTestBook(t, f.BookService, "Test Book", series.ID)

	// Find covers for the book
	covers, err := f.CoverService.FindForBook(context.Background(), book)
	require.NoError(t, err)
	require.Len(t, covers, 0)
}

func TestCoverService_GetByID_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	cover := &models.Cover{
		Text:        "Test Cover",
		SeriesID:    series.ID,
		Artist:      "Test Artist",
		Filename:    "test.jpg",
		IssueNumber: 1,
		Publication: "2023-01-01",
	}

	err := f.CoverService.Upsert(context.Background(), cover)
	require.NoError(t, err)

	// Retrieve the cover by ID
	retrievedCover, err := f.CoverService.GetByID(context.Background(), cover.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedCover)
	require.Equal(t, cover.ID, retrievedCover.ID)
	require.Equal(t, cover.Text, retrievedCover.Text)
}

func TestCoverService_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Try to get a non-existent cover
	cover, err := f.CoverService.GetByID(context.Background(), 999)
	require.Error(t, err)
	require.Nil(t, cover)
}

func TestCoverService_ListAll_ReturnsPaginatedCovers(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create multiple covers
	for i := 1; i <= 3; i++ {
		cover := &models.Cover{
			Text:        "Cover " + string(rune('A'+i-1)),
			SeriesID:    series.ID,
			Artist:      "Artist",
			Filename:    "file.jpg",
			IssueNumber: i,
			Publication: "2023-01-01",
		}
		err := f.CoverService.Upsert(context.Background(), cover)
		require.NoError(t, err)
	}

	// List all covers with pagination
	covers, total, err := f.CoverService.ListAll(context.Background(), 0, 2)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, covers, 2)
}

func TestCoverService_ListAll_Empty(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// List all covers when none exist
	covers, total, err := f.CoverService.ListAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Len(t, covers, 0)
}

func TestCoverService_FindAllBySeries_ReturnsCoversList(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create multiple covers for the same series
	for i := 1; i <= 2; i++ {
		cover := &models.Cover{
			Text:        "Cover " + string(rune('A'+i-1)),
			SeriesID:    series.ID,
			Artist:      "Artist",
			Filename:    "file.jpg",
			IssueNumber: i,
			Publication: "2023-01-01",
		}
		err := f.CoverService.Upsert(context.Background(), cover)
		require.NoError(t, err)
	}

	// Find all covers for the series
	covers, err := f.CoverService.FindAllBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.Len(t, covers, 2)
}

func TestCoverService_FindAllBySeries_Empty(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Find all covers for series with no covers
	covers, err := f.CoverService.FindAllBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.Len(t, covers, 0)
}

func TestCoverService_FindFirstBySeries_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create multiple covers
	for i := 3; i >= 1; i-- {
		cover := &models.Cover{
			Text:        "Cover",
			SeriesID:    series.ID,
			Artist:      "Artist",
			Filename:    "file.jpg",
			IssueNumber: i,
			Publication: "2023-01-01",
		}
		err := f.CoverService.Upsert(context.Background(), cover)
		require.NoError(t, err)
	}

	// Find first cover (should be ordered by issue number)
	cover, err := f.CoverService.FindFirstBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.NotNil(t, cover)
	require.Equal(t, 1, cover.IssueNumber)
}

func TestCoverService_FindFirstBySeries_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Try to find first cover when none exist
	cover, err := f.CoverService.FindFirstBySeries(context.Background(), series.ID)
	require.Error(t, err)
	require.Nil(t, cover)
}

func TestCoverService_FindBySeriesName_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create a cover for the series
	cover := &models.Cover{
		Text:        "Test Cover",
		SeriesID:    series.ID,
		Artist:      "Artist",
		Filename:    "file.jpg",
		IssueNumber: 1,
		Publication: "2023-01-01",
	}
	err := f.CoverService.Upsert(context.Background(), cover)
	require.NoError(t, err)

	// Find cover by series name
	foundCover, err := f.CoverService.FindBySeriesName(context.Background(), series.Name)
	require.NoError(t, err)
	require.NotNil(t, foundCover)
	require.Equal(t, cover.SeriesID, foundCover.SeriesID)
}

func TestCoverService_FindBySeriesName_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Try to find cover for series that exists but has no covers
	cover, err := f.CoverService.FindBySeriesName(context.Background(), series.Name)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
	require.Nil(t, cover)
}

func TestCoverService_FindBySeriesName_SeriesHasNoCover(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Try to find cover for series with no covers
	cover, err := f.CoverService.FindBySeriesName(context.Background(), series.Name)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
	require.Nil(t, cover)
}

func TestCoverService_UpsertFromAPI_CreatesNewCover(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create test cover and issue from scan API
	apiCover := CreateTestCover(t, f.CoverService, series.ID, 100)
	require.NotNil(t, apiCover)
	require.Greater(t, apiCover.ID, int64(0))
}
