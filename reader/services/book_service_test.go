package services

import (
	"context"
	"testing"

	"github.com/chooban/progger/reader/models"
	"github.com/stretchr/testify/require"
)

func TestBookService_ListBySeries_ReturnsPaginatedBooks(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	_ = CreateTestBook(t, f.BookService, "First Book", series.ID)
	_ = CreateTestBook(t, f.BookService, "Second Book", series.ID)

	books, total, err := f.BookService.ListBySeries(context.Background(), series.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, books, 2)
}

func TestBookService_ListBySeries_Empty(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	books, total, err := f.BookService.ListBySeries(context.Background(), series.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Len(t, books, 0)
}

func TestBookService_GetByID_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "First Book", series.ID)

	foundBook, err := f.BookService.GetByID(context.Background(), book.ID)
	require.NoError(t, err)
	require.NotNil(t, foundBook)
	require.Equal(t, book.ID, foundBook.ID)
	require.NotNil(t, foundBook.Episodes)
}

func TestBookService_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	book, err := f.BookService.GetByID(context.Background(), 999)
	require.Error(t, err)
	require.Nil(t, book)
}

func TestBookService_Upsert_InsertsNewBook(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	book := &models.Book{
		SeriesID:   series.ID,
		Name:       "New Book",
		Number:     1,
		Status:     "READY",
		PageCount:  32,
		FirstIssue: 1,
		LastIssue:  1,
	}

	err := f.BookService.Upsert(context.Background(), book)
	require.NoError(t, err)
	require.Greater(t, book.ID, int64(0))
	require.NotEmpty(t, book.CreatedAt)
	require.NotEmpty(t, book.UpdatedAt)
}

func TestBookService_UpsertEpisodes_InsertsEpisodes(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "First Book", series.ID)

	episodes := []*models.Episode{
		{
			BookID:      book.ID,
			Filename:    "episode1.cbz",
			IssueNumber: 1,
			Title:       "First Episode",
			Part:        0,
			PageFrom:    1,
			PageTo:      16,
		},
		{
			BookID:      book.ID,
			Filename:    "episode2.cbz",
			IssueNumber: 2,
			Title:       "Second Episode",
			Part:        0,
			PageFrom:    17,
			PageTo:      32,
		},
	}

	err := f.BookService.UpsertEpisodes(context.Background(), episodes)
	require.NoError(t, err)

	for _, ep := range episodes {
		require.Greater(t, ep.ID, int64(0))
		require.Equal(t, book.ID, ep.BookID)
	}
}

func TestBookService_ListAll_ReturnsPaginatedBooks(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Create test data across multiple series
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series1 := CreateTestSeries(t, f.SeriesService, lib.ID)
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series 2")

	_ = CreateTestBook(t, f.BookService, "First", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Second", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Third", series2.ID)

	books, total, err := f.BookService.ListAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, books, 3)
}

func TestBookService_ListAll_Empty(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	books, total, err := f.BookService.ListAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, books)
}

func TestBookService_ListByLibraryID_ReturnsPaginatedBooks(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Create two libraries
	lib1 := CreateTestLibrary(t, f.LibraryService, "Library 1")
	lib2 := CreateTestLibrary(t, f.LibraryService, "Library 2")

	// Create series in different libraries
	series1 := CreateTestSeriesWithName(t, f.SeriesService, lib1.ID, "Series A")
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib2.ID, "Series B")

	// Add books to each
	_ = CreateTestBook(t, f.BookService, "First", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Second", series2.ID)

	// Test filtering by library 1
	books, total, err := f.BookService.ListByLibraryID(context.Background(), lib1.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, books, 1)
	require.Equal(t, series1.ID, books[0].SeriesID)
}

func TestBookService_ListByLibraryID_EmptyLibrary(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Create library with no books
	lib := CreateTestLibrary(t, f.LibraryService, "Empty Library")

	books, total, err := f.BookService.ListByLibraryID(context.Background(), lib.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, books)
}

func TestBookService_ListBySeries_OnlyReturnsBooksForSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series1 := CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series A")
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series B")
	series3 := CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series C")

	_ = CreateTestBook(t, f.BookService, "First", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Second", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Third", series2.ID)
	_ = CreateTestBook(t, f.BookService, "Fourth", series2.ID)
	_ = CreateTestBook(t, f.BookService, "Fifth", series3.ID)

	// Test filtering by series 1 - should return 2 books only
	books1, total1, err := f.BookService.ListBySeries(context.Background(), series1.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total1)
	require.Len(t, books1, 2)
	for _, book := range books1 {
		require.Equal(t, series1.ID, book.SeriesID)
	}

	// Test filtering by series 2 - should return 2 books only
	books2, total2, err := f.BookService.ListBySeries(context.Background(), series2.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total2)
	require.Len(t, books2, 2)
	for _, book := range books2 {
		require.Equal(t, series2.ID, book.SeriesID)
	}

	// Test filtering by series 3 - should return 1 book only
	books3, total3, err := f.BookService.ListBySeries(context.Background(), series3.ID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 1, total3)
	require.Len(t, books3, 1)
	require.Equal(t, series3.ID, books3[0].SeriesID)
}

func TestBookService_CountBySeries_ReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Create multiple books for the series
	for i := 1; i <= 3; i++ {
		book := &models.Book{
			SeriesID: series.ID,
			Name:     "Book " + string(rune('A'+i-1)),
		}
		err := f.BookService.Upsert(context.Background(), book)
		require.NoError(t, err)
	}

	// Count books in the series
	count, err := f.BookService.CountBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestBookService_CountBySeries_EmptySeriesReturnsZero(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	// Count books in empty series
	count, err := f.BookService.CountBySeries(context.Background(), series.ID)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestBookService_ListLatest_ReturnsPaginatedBooks(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	_ = CreateTestBook(t, f.BookService, "First Book", series.ID)
	_ = CreateTestBook(t, f.BookService, "Second Book", series.ID)

	books, total, err := f.BookService.ListLatest(context.Background(), nil, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, books, 2)
}

func TestBookService_ListLatest_FiltersByLibrary(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib1 := CreateTestLibrary(t, f.LibraryService, "Library 1")
	lib2 := CreateTestLibrary(t, f.LibraryService, "Library 2")

	series1 := CreateTestSeries(t, f.SeriesService, lib1.ID)
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib2.ID, "Series 2")

	_ = CreateTestBook(t, f.BookService, "Book 1", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Book 2", series2.ID)

	books, total, err := f.BookService.ListLatest(context.Background(), []int64{lib1.ID}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, books, 1)
}

func TestBookService_GetNextBook_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	first := CreateTestBook(t, f.BookService, "First", series.ID)
	second := CreateTestBook(t, f.BookService, "Second", series.ID)

	// Set sequential numbers for ordering
	_, err := f.DB.ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", first.ID)
	require.NoError(t, err)
	_, err = f.DB.ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", second.ID)
	require.NoError(t, err)

	next, err := f.BookService.GetNextBook(context.Background(), first.ID)
	require.NoError(t, err)
	require.Equal(t, second.ID, next.ID)
}

func TestBookService_GetNextBook_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	last := CreateTestBook(t, f.BookService, "Last Book", series.ID)

	_, err := f.BookService.GetNextBook(context.Background(), last.ID)
	require.Error(t, err)
}

func TestBookService_GetPreviousBook_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	first := CreateTestBook(t, f.BookService, "First", series.ID)
	second := CreateTestBook(t, f.BookService, "Second", series.ID)

	// Set sequential numbers for ordering
	_, err := f.DB.ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", first.ID)
	require.NoError(t, err)
	_, err = f.DB.ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", second.ID)
	require.NoError(t, err)

	prev, err := f.BookService.GetPreviousBook(context.Background(), second.ID)
	require.NoError(t, err)
	require.Equal(t, first.ID, prev.ID)
}

func TestBookService_GetPreviousBook_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	first := CreateTestBook(t, f.BookService, "First Book", series.ID)

	_, err := f.BookService.GetPreviousBook(context.Background(), first.ID)
	require.Error(t, err)
}

func TestBookService_ListOnDeck_ReturnsFirstBookPerSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series1 := CreateTestSeries(t, f.SeriesService, lib.ID)
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series B")

	// Create multiple books for each series
	_ = CreateTestBook(t, f.BookService, "Book 1.1", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Book 1.2", series1.ID)
	_ = CreateTestBook(t, f.BookService, "Book 2.1", series2.ID)

	// List on deck books (should be first book per series)
	books, total, err := f.BookService.ListOnDeck(context.Background(), []int64{}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, books, 2)

	// Verify we got the first books
	seriesIDs := make(map[int64]bool)
	for _, book := range books {
		require.False(t, seriesIDs[book.SeriesID], "Should only get one book per series")
		seriesIDs[book.SeriesID] = true
	}
}

func TestBookService_ListOnDeck_FiltersbyLibrary(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib1 := CreateTestLibrary(t, f.LibraryService, "Library 1")
	lib2 := CreateTestLibrary(t, f.LibraryService, "Library 2")

	series1 := CreateTestSeries(t, f.SeriesService, lib1.ID)
	series2 := CreateTestSeriesWithName(t, f.SeriesService, lib2.ID, "Series B")

	// Create books in both libraries
	CreateTestBook(t, f.BookService, "Book 1", series1.ID)
	CreateTestBook(t, f.BookService, "Book 2", series2.ID)

	// List on deck books for lib1 only
	books, total, err := f.BookService.ListOnDeck(context.Background(), []int64{lib1.ID}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, books, 1)
	require.Equal(t, series1.ID, books[0].SeriesID)
}

func TestBookService_GetForThumbnail_ReturnsBook(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Test Book", series.ID)

	// Create an episode for the book
	CreateTestEpisode(t, f.BookService, book.ID, 1, 0)

	// Get book for thumbnail
	result, err := f.BookService.GetForThumbnail(context.Background(), book.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, book.ID, result.Book.ID)
}

func TestBookService_GetForThumbnail_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Try to get non-existent book
	result, err := f.BookService.GetForThumbnail(context.Background(), 999)
	require.Error(t, err)
	require.Nil(t, result)
}
