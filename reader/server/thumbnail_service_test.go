package server

import (
	"context"
	"errors"
	"image"
	"testing"

	scanApi "github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/database"
	"github.com/stretchr/testify/require"
)

func thumbnailCoverBuilder(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, 100, 100)), nil
}

func thumbnailSeriesBuilder(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, 200, 150)), nil
}

func TestResizeImage_Landscape(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	resized := resizeImage(img, 300)

	bounds := resized.Bounds()
	require.Equal(t, 300, bounds.Dx())
	require.Equal(t, 225, bounds.Dy())
}

func TestResizeImage_Portrait(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 600, 800))
	resized := resizeImage(img, 300)

	bounds := resized.Bounds()
	require.Equal(t, 225, bounds.Dx())
	require.Equal(t, 300, bounds.Dy())
}

func TestResizeImage_Square(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	resized := resizeImage(img, 300)

	bounds := resized.Bounds()
	require.Equal(t, 300, bounds.Dx())
	require.Equal(t, 300, bounds.Dy())
}

func TestResizeImage_SmallUpscales(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	resized := resizeImage(img, 300)

	bounds := resized.Bounds()
	require.Equal(t, 300, bounds.Dx())
	require.Equal(t, 300, bounds.Dy())
}

func TestResizeImage_MinDimClamp(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	resized := resizeImage(img, 1)

	bounds := resized.Bounds()
	require.Equal(t, 1, bounds.Dx())
	require.Equal(t, 1, bounds.Dy())
}

func TestThumbnailService_GetCoverThumbnail(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	cover := &database.Cover{
		ID:       1,
		Filename: "test-cover.pdf",
	}

	data, err := svc.GetCoverThumbnail(context.Background(), cover)
	require.NoError(t, err)
	require.Greater(t, len(data), 2)
	require.Equal(t, []byte{0xff, 0xd8}, data[:2])
}

func TestThumbnailService_GetCoverThumbnail_NilCover(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	data, err := svc.GetCoverThumbnail(context.Background(), nil)
	require.Error(t, err)
	require.Empty(t, data)
	require.Contains(t, err.Error(), "cover is nil")
}

func TestThumbnailService_GetCoverThumbnail_EmptyFilename(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	cover := &database.Cover{
		ID:       1,
		Filename: "",
	}

	data, err := svc.GetCoverThumbnail(context.Background(), cover)
	require.NoError(t, err)
	require.Equal(t, []byte("no thumbnail available"), data)
}

func TestThumbnailService_GetCoverThumbnail_CacheHit(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	cover := &database.Cover{
		ID:       1,
		Filename: "test-cover.pdf",
	}

	data1, err := svc.GetCoverThumbnail(context.Background(), cover)
	require.NoError(t, err)

	data2, err := svc.GetCoverThumbnail(context.Background(), cover)
	require.NoError(t, err)

	require.Equal(t, data1, data2)
}

func TestThumbnailService_GetCoverThumbnail_BuilderError(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, func(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
		return nil, errors.New("mock error")
	})
	svc := NewThumbnailService(pageSvc)

	cover := &database.Cover{
		ID:       1,
		Filename: "test-cover.pdf",
	}

	_, err := svc.GetCoverThumbnail(context.Background(), cover)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to build page")
}

func TestThumbnailService_GetPageThumbnail(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Test Book", series.ID)
	CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)

	data, err := svc.GetPageThumbnail(context.Background(), book.ID, 1)
	require.NoError(t, err)
	require.Greater(t, len(data), 2)
	require.Equal(t, []byte{0xff, 0xd8}, data[:2])
}

func TestThumbnailService_GetPageThumbnail_BookNotFound(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	_, err := svc.GetPageThumbnail(context.Background(), 99999, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get book")
}

func TestThumbnailService_GetPageThumbnail_PageOutOfRange(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Test Book", series.ID)
	CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)

	_, err := svc.GetPageThumbnail(context.Background(), book.ID, 999)
	require.Error(t, err)
	require.Contains(t, err.Error(), "page")
}

func TestThumbnailService_GetPageThumbnail_CacheHit(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)
	pageSvc := NewPageService(f.BookRepo, nil, nil, thumbnailCoverBuilder)
	svc := NewThumbnailService(pageSvc)

	lib := CreateTestLibrary(t, f.LibraryRepo, "Test Library")
	series := CreateTestSeries(t, f.SeriesRepo, lib.ID)
	book := CreateTestBook(t, f.BookRepo, "Test Book", series.ID)
	CreateTestEpisode(t, f.BookRepo, book.ID, 1, 1)

	data1, err := svc.GetPageThumbnail(context.Background(), book.ID, 1)
	require.NoError(t, err)

	data2, err := svc.GetPageThumbnail(context.Background(), book.ID, 1)
	require.NoError(t, err)

	require.Equal(t, data1, data2)
}

func TestThumbnailService_GetSeriesThumbnail(t *testing.T) {
	t.Parallel()

	svc := &ThumbnailService{
		seriesCoverBuilder: thumbnailSeriesBuilder,
	}

	cover := &database.Cover{
		Filename: "test-series-cover.pdf",
	}

	data, err := svc.GetSeriesThumbnail(context.Background(), cover, "Judge Dredd")
	require.NoError(t, err)
	require.Greater(t, len(data), 2)
	require.Equal(t, []byte{0xff, 0xd8}, data[:2])
}

func TestThumbnailService_GetSeriesThumbnail_CacheHit(t *testing.T) {
	t.Parallel()

	svc := &ThumbnailService{
		seriesCoverBuilder: thumbnailSeriesBuilder,
	}

	cover := &database.Cover{
		Filename: "test-series-cover.pdf",
	}

	data1, err := svc.GetSeriesThumbnail(context.Background(), cover, "Rogue Trooper")
	require.NoError(t, err)

	data2, err := svc.GetSeriesThumbnail(context.Background(), cover, "Rogue Trooper")
	require.NoError(t, err)

	require.Equal(t, data1, data2)
}

func TestThumbnailService_GetSeriesThumbnail_BuilderError(t *testing.T) {
	t.Parallel()

	svc := &ThumbnailService{
		seriesCoverBuilder: func(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
			return nil, errors.New("mock error")
		},
	}

	cover := &database.Cover{
		Filename: "test-series-cover.pdf",
	}

	_, err := svc.GetSeriesThumbnail(context.Background(), cover, "Strontium Dog")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to build page")
}
