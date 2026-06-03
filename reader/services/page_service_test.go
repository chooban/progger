package services

import (
	"context"
	"image"
	"testing"

	scanApi "github.com/chooban/progger/scan/api"
	"github.com/stretchr/testify/require"
)

func mockCoverBuilder(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {

	return image.NewRGBA(image.Rect(0, 0, 100, 100)), nil
}

func mockPageBuilder(ctx context.Context, page scanApi.ExportPage) (*[]byte, error) {
	// Return valid JPEG magic bytes for testing
	jpegData := []byte{0xff, 0xd8, 0xff, 0xe0}
	return &jpegData, nil
}

func TestPageService_GetPages_ReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookService, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Book", series.ID)
	ep1 := CreateTestEpisode(t, f.BookService, book.ID, 1, 1)
	ep2 := CreateTestEpisode(t, f.BookService, book.ID, 1, 2)

	// Update episodes to create a total of 32 pages (16 pages each)
	_, err := f.DB.ExecContext(context.Background(),
		"UPDATE episodes SET page_from=?, page_to=? WHERE id=?",
		1, 16, ep1.ID)
	require.NoError(t, err)
	_, err = f.DB.ExecContext(context.Background(),
		"UPDATE episodes SET page_from=?, page_to=? WHERE id=?",
		17, 32, ep2.ID)
	require.NoError(t, err)

	pages, err := pageService.GetPages(context.Background(), book.ID)
	require.NoError(t, err)
	require.Len(t, pages, 32)
}

func TestPageService_GetPages_EmptyEpisodes(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookService, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Book", series.ID)

	pages, err := pageService.GetPages(context.Background(), book.ID)
	require.NoError(t, err)
	require.Empty(t, pages)
}

func TestPageService_GetPage_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookService, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Book", series.ID)
	CreateTestEpisode(t, f.BookService, book.ID, 1, 1)

	_, err := pageService.GetPage(context.Background(), book.ID, 999)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestPageService_GetPage_CalculatesRelativePage(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookService, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Book", series.ID)
	ep := CreateTestEpisode(t, f.BookService, book.ID, 1, 1)

	// Test second page of first episode
	_, err := f.DB.ExecContext(context.Background(),
		"UPDATE episodes SET page_from=1, page_to=10 WHERE id=?",
		ep.ID)
	require.NoError(t, err)

	data, err := pageService.GetPage(context.Background(), book.ID, 2)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data), 2, "Should have at least 2 bytes for JPEG magic number")
	require.Equal(t, []byte{0xff, 0xd8}, data[:2], "Should have JPEG magic bytes")
}

func TestPageService_GetPage_CachesResult(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookService, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)
	book := CreateTestBook(t, f.BookService, "Book", series.ID)
	_ = CreateTestEpisode(t, f.BookService, book.ID, 1, 1)

	// Test caching
	data1, err := pageService.GetPage(context.Background(), book.ID, 1)
	require.NoError(t, err)

	data2, err := pageService.GetPage(context.Background(), book.ID, 1)
	require.NoError(t, err)

	require.Equal(t, string(data1), string(data2))
	require.GreaterOrEqual(t, len(data1), 2, "Should have at least 2 bytes for JPEG magic number")
	require.Equal(t, []byte{0xff, 0xd8}, data1[:2], "Should have JPEG magic bytes")
}
