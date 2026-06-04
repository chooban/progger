package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageService_GetPages_ReturnsCorrectCount(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookRepo, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Book", series.ID)
	ep1 := CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)
	ep2 := CreateTestEpisode(t, f.BookRepo, book.ID, 1, 2)

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
	pageService := NewPageService(f.BookRepo, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Book", series.ID)

	pages, err := pageService.GetPages(context.Background(), book.ID)
	require.NoError(t, err)
	require.Empty(t, pages)
}

func TestPageService_GetPage_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookRepo, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Book", series.ID)
	ep := CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)
	_, err := f.DB.ExecContext(context.Background(),
		"UPDATE episodes SET page_from=?, page_to=? WHERE id=?",
		1, 16, ep.ID)
	require.NoError(t, err)

	_, err = pageService.GetPage(context.Background(), book.ID, 999)
	require.Error(t, err)
}

func TestPageService_GetPageByImage_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageService := NewPageService(f.BookRepo, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Book", series.ID)
	ep := CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)
	_, err := f.DB.ExecContext(context.Background(),
		"UPDATE episodes SET page_from=?, page_to=? WHERE id=?",
		1, 16, ep.ID)
	require.NoError(t, err)

	_, err = pageService.GetPageImage(context.Background(), book.ID, 999)
	require.Error(t, err)
}

func TestPageService_FindEpisodeForPage(t *testing.T) {
	f := setupTestFixture(t)
	pageService := NewPageService(f.BookRepo, nil, mockPageBuilder, mockCoverBuilder)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Book", series.ID)
	CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)
	CreateTestEpisode(t, f.BookRepo, book.ID, 1, 2)

	pageByID, err := pageService.bookSer.GetByID(context.Background(), book.ID)
	require.NoError(t, err)
	require.NotNil(t, pageByID)

	ep, pageInEp, err := pageService.FindEpisodeForPage(pageByID.Episodes, 1)
	require.NoError(t, err)
	require.NotNil(t, ep)
	require.Equal(t, 0, pageInEp)
}
