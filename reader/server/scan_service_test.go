package server

import (
	"context"
	"testing"

	scanApi "github.com/chooban/progger/scan/api"
	"github.com/stretchr/testify/require"
)

func TestScanService_ScanDirectories_CreatesSeriesAndBooks(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Create mock scanner returning test data
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			return []scanApi.Issue{
				{
					Publication: "2000 AD",
					IssueNumber: 1,
					Episodes: []*scanApi.Episode{
						{
							Series:    "Test Series",
							Title:     "Test Episode",
							Part:      0,
							FirstPage: 1,
							LastPage:  32,
						},
					},
					Filename: "test.pdf",
					Cover:    scanApi.Cover{Series: "Test Series", Artist: "Test Artist", Filename: "cover.pdf"},
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Verify library was created
	libs, err := f.LibraryRepo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, libs, 1)
	require.Equal(t, "Test Library", libs[0].Name)

	// Verify series was created
	series, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Test Series")
	require.NoError(t, err)
	require.NotNil(t, series)

	// Verify book was created
	seriesBooks, total, err := f.BookRepo.ListBySeries(context.Background(), series.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, seriesBooks, 1)
	require.Equal(t, 1, total)
	require.Equal(t, "Test Episode", seriesBooks[0].Name)
}

func TestScanService_ScanDirectories_MultipleSeriesAreCreated(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Mock with multiple series
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			return []scanApi.Issue{
				{
					Publication: "2000 AD",
					IssueNumber: 1,
					Episodes:    []*scanApi.Episode{{Series: "Series A", Title: "Episode A"}},
					Filename:    "a.pdf",
					Cover:       scanApi.Cover{Series: "Series A"},
				},
				{
					Publication: "2000 AD",
					IssueNumber: 2,
					Episodes:    []*scanApi.Episode{{Series: "Series B", Title: "Episode B"}},
					Filename:    "b.pdf",
					Cover:       scanApi.Cover{Series: "Series B"},
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Verify both series were created
	seriesA, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Series A")
	require.NoError(t, err)
	require.NotNil(t, seriesA)

	seriesB, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Series B")
	require.NoError(t, err)
	require.NotNil(t, seriesB)

	// Verify each series has one book
	booksA, _, err := f.BookRepo.ListBySeries(context.Background(), seriesA.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, booksA, 1)

	booksB, _, err := f.BookRepo.ListBySeries(context.Background(), seriesB.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, booksB, 1)
}

func TestScanService_ScanDirectories_EmptyIssuesAreSkipped(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Mock scanner returns empty issue
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			return []scanApi.Issue{
				{
					Filename:    "invalid.pdf",
					Publication: "2000 AD",
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Verify no books were created
	books, total, err := f.BookRepo.ListAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Len(t, books, 0)
}

func TestScanService_ScanDirectories_SeriesBookCountUpdated(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Mock scanner returning multiple books for same series
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			return []scanApi.Issue{
				{
					Publication: "2000 AD",
					IssueNumber: 1,
					Episodes:    []*scanApi.Episode{{Series: "Test Series", Title: "Book A", Part: 0, FirstPage: 1, LastPage: 32}},
					Filename:    "bookA.pdf",
				},
				{
					Publication: "2000 AD",
					IssueNumber: 2,
					Episodes:    []*scanApi.Episode{{Series: "Test Series", Title: "Book B", Part: 0, FirstPage: 1, LastPage: 32}},
					Filename:    "bookB.pdf",
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Verify series book count is updated
	series, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Test Series")
	require.NoError(t, err)
	require.Equal(t, 2, series.BookCount)
}

func TestScanService_ScanDirectories_EpisodePageCounts(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	// Mock with multi-episode book (same title groups into one book)
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			return []scanApi.Issue{
				{
					Publication: "2000 AD",
					IssueNumber: 1,
					Episodes: []*scanApi.Episode{
						{Series: "Test Series", Title: "Full Story", Part: 0, FirstPage: 1, LastPage: 16},
						{Series: "Test Series", Title: "Full Story", Part: 1, FirstPage: 17, LastPage: 32},
					},
					Filename: "multi-chapter.pdf",
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Verify book has 32 pages (16+16)
	series, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Test Series")
	require.NoError(t, err)
	books, _, err := f.BookRepo.ListBySeries(context.Background(), series.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, books, 1)
	require.Equal(t, 32, books[0].PageCount)
}

func TestScanService_ScanDirectories_IdempotentOnSecondScan(t *testing.T) {
	t.Parallel()

	f := setupTestFixture(t)

	callCount := 0
	// Mock with multi-episode book
	mockScanner := &MockScanner{
		DirFunc: func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
			callCount++
			return []scanApi.Issue{
				{
					Publication: "2000 AD",
					IssueNumber: 1,
					Episodes: []*scanApi.Episode{
						{Series: "Test Series", Title: "Full Story", Part: 0, FirstPage: 1, LastPage: 16},
						{Series: "Test Series", Title: "Full Story", Part: 1, FirstPage: 17, LastPage: 32},
					},
					Filename: "multi-chapter.pdf",
				},
			}, nil
		},
	}

	cfgPkg := &Config{
		LibraryName:     "Test Library",
		ScanDirectories: []string{"/test/path"},
	}

	scanSer := NewScanService(f.DB, cfgPkg, f.LibraryRepo, f.SeriesRepo, f.BookRepo, f.CoverRepo, mockScanner)

	// First scan
	err := scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Count entities after first scan
	seriesA, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Test Series")
	require.NoError(t, err)
	require.NotNil(t, seriesA)

	booksA, totalA, err := f.BookRepo.ListBySeries(context.Background(), seriesA.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, booksA, 1)

	bookA, err := f.BookRepo.GetByID(context.Background(), booksA[0].ID)
	require.NoError(t, err)
	require.NotNil(t, bookA.Episodes)
	initialEpisodeCount := len(bookA.Episodes)

	// Second scan with same data
	err = scanSer.ScanDirectories(context.Background())
	require.NoError(t, err)

	// Count entities after second scan
	seriesB, err := f.SeriesRepo.MaybeGetByName(context.Background(), "Test Series")
	require.NoError(t, err)
	require.NotNil(t, seriesB)
	require.Equal(t, seriesA.ID, seriesB.ID)
	require.Equal(t, seriesA.Name, seriesB.Name)
	require.Equal(t, seriesA.BookCount, seriesB.BookCount)

	booksB, totalB, err := f.BookRepo.ListBySeries(context.Background(), seriesB.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, booksB, 1)
	require.Equal(t, totalA, totalB)

	bookB, err := f.BookRepo.GetByID(context.Background(), booksB[0].ID)
	require.NoError(t, err)
	require.NotNil(t, bookB.Episodes)
	require.Equal(t, len(bookA.Episodes), len(bookB.Episodes))
	require.Equal(t, initialEpisodeCount, len(bookB.Episodes))

	// Verify no new IDs were created on second scan
	require.Equal(t, booksA[0].ID, booksB[0].ID, "Book ID should be identical after rescanning")

	// Verify episode IDs remain unchanged
	require.Equal(t, len(bookA.Episodes), len(bookB.Episodes), "Episode count should be identical after rescanning")
	if len(bookA.Episodes) > 0 {
		episodeIDsA := make(map[int64]bool)
		for _, ep := range bookA.Episodes {
			episodeIDsA[ep.ID] = true
		}
		episodeIDsB := make(map[int64]bool)
		for _, ep := range bookB.Episodes {
			episodeIDsB[ep.ID] = true
		}
		require.Equal(t, episodeIDsA, episodeIDsB, "Episode IDs should be identical after rescanning")
	}

	// Ensure scanner was called twice
	require.Equal(t, 2, callCount)
}
