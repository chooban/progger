package services

import (
	"context"
	"testing"
	"time"

	"github.com/chooban/progger/reader/models"
	"github.com/stretchr/testify/require"
)

func TestSeriesService_List_ReturnsPaginatedSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")

	for i := 1; i <= 5; i++ {
		seriesEntity := &models.Series{
			LibraryID: lib.ID,
			Name:      "Test Series " + string(rune('A'+i-1)),
			BookCount: 0,
			CreatedAt: FormatTime(time.Now()),
			UpdatedAt: FormatTime(time.Now()),
		}
		_, err := f.DB.ExecContext(context.Background(),
			"INSERT INTO series (library_id, name, book_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			seriesEntity.LibraryID, seriesEntity.Name, seriesEntity.BookCount,
			seriesEntity.CreatedAt, seriesEntity.UpdatedAt)
		require.NoError(t, err)
	}

	allSeries, total, err := f.SeriesService.List(context.Background(), 0, 2)
	require.NoError(t, err)
	require.Equal(t, 5, total)
	require.Len(t, allSeries, 2)
}

func TestSeriesService_GetByID_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	testSeries := CreateTestSeries(t, f.SeriesService, lib.ID)

	foundSeries, err := f.SeriesService.GetByID(context.Background(), testSeries.ID)
	require.NoError(t, err)
	require.NotNil(t, foundSeries)
	require.Equal(t, testSeries.ID, foundSeries.ID)
	require.Equal(t, testSeries.Name, foundSeries.Name)
}

func TestSeriesService_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	series, err := f.SeriesService.GetByID(context.Background(), 999)
	require.Error(t, err)
	require.Nil(t, series)
}

func TestSeriesService_GetByName_Found(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	testSeries := CreateTestSeries(t, f.SeriesService, lib.ID)

	foundSeries, err := f.SeriesService.MaybeGetByName(context.Background(), testSeries.Name)
	require.NoError(t, err)
	require.NotNil(t, foundSeries)
	require.Equal(t, testSeries.Name, foundSeries.Name)
}

func TestSeriesService_GetByName_NotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	series, err := f.SeriesService.MaybeGetByName(context.Background(), "Non Existent")
	require.Nil(t, err)
	require.Nil(t, series)
}

func TestSeriesService_Upsert_InsertsNewSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")

	series := &models.Series{
		LibraryID: lib.ID,
		Name:      "New Series",
		BookCount: 0,
	}
	err := f.SeriesService.Upsert(context.Background(), series)
	require.NoError(t, err)
	require.Greater(t, series.ID, int64(0))
}

func TestSeriesService_Upsert_Conflict(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")

	series := &models.Series{
		LibraryID: lib.ID,
		Name:      "New Series",
		BookCount: 0,
	}
	err := f.SeriesService.Upsert(context.Background(), series)
	require.NoError(t, err)
	require.Greater(t, series.ID, int64(0))

	duplicateSeries := &models.Series{
		LibraryID: lib.ID,
		Name:      "New Series",
		BookCount: 0,
	}
	err = f.SeriesService.Upsert(context.Background(), duplicateSeries)
	require.NoError(t, err)
	require.Greater(t, duplicateSeries.ID, int64(0))
	require.Equal(t, series.ID, duplicateSeries.ID)
}

func TestSeriesService_IncrementBookCount_Increments(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	series := CreateTestSeries(t, f.SeriesService, lib.ID)

	err := f.SeriesService.SetBookCount(context.Background(), series.ID, 0)
	require.NoError(t, err)

	updatedSeries, err := f.SeriesService.GetByID(context.Background(), series.ID)
	require.NoError(t, err)
	require.Equal(t, 0, updatedSeries.BookCount)
}

func TestSeriesService_ListRecentlyAdded_ReturnsRecentSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	_ = CreateTestSeries(t, f.SeriesService, lib.ID)
	_ = CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series B")

	series, total, err := f.SeriesService.ListRecentlyAdded(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, series, 2)
	require.Equal(t, 2, total)
}

func TestSeriesService_ListRecentlyUpdated_ReturnsRecentSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	_ = CreateTestSeries(t, f.SeriesService, lib.ID)
	_ = CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series B")

	series, total, err := f.SeriesService.ListRecentlyUpdated(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, series, 2)
	require.Equal(t, 2, total)
}

func TestSeriesService_ListLatest_ReturnsPaginatedSeries(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib := CreateTestLibrary(t, f.LibraryService, "Test Library")
	_ = CreateTestSeries(t, f.SeriesService, lib.ID)
	_ = CreateTestSeriesWithName(t, f.SeriesService, lib.ID, "Series B")

	series, total, err := f.SeriesService.ListLatest(context.Background(), nil, 0, 10)
	require.NoError(t, err)
	require.Len(t, series, 2)
	require.Equal(t, 2, total)
}

func TestSeriesService_ListLatest_FiltersByLibrary(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	lib1 := CreateTestLibrary(t, f.LibraryService, "Library 1")
	lib2 := CreateTestLibrary(t, f.LibraryService, "Library 2")

	_ = CreateTestSeries(t, f.SeriesService, lib1.ID)
	_ = CreateTestSeriesWithName(t, f.SeriesService, lib2.ID, "Series B")

	series, total, err := f.SeriesService.ListLatest(context.Background(), []int64{lib1.ID}, 0, 10)
	require.NoError(t, err)
	require.Len(t, series, 1)
	require.Equal(t, 1, total)
}
