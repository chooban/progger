package server

import (
	"context"
	"fmt"
	"image"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/chooban/progger/database"
	scanApi "github.com/chooban/progger/scan/api"
	"github.com/gin-gonic/gin"
)

// Mock builders for testing
func mockCoverBuilder(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, 100, 100)), nil
}

func mockPageBuilder(ctx context.Context, page scanApi.ExportPage) (*[]byte, error) {
	a := make([]byte, 100)
	return &a, nil
}

// Router and server setup
func setupTestRouter(handlers *Handlers) *gin.Engine {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	ConfigureRoutes(router, handlers)
	return router
}

func createTestServer(t *testing.T, handlers *Handlers) *httptest.Server {
	t.Helper()
	router := setupTestRouter(handlers)
	return httptest.NewServer(router)
}

// Test database management
var (
	testDBMutex sync.Mutex
	testDBMap   = make(map[*testing.T]*database.DB)
)

func createTestHandlers(t *testing.T) *Handlers {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	testDBMutex.Lock()
	testDBMap[t] = db
	testDBMutex.Unlock()

	t.Cleanup(func() {
		testDBMutex.Lock()
		delete(testDBMap, t)
		testDBMutex.Unlock()
		db.Close()
	})

	libraryRepo := database.NewLibraryRepo(db)
	seriesRepo := database.NewSeriesRepo(db)
	bookRepo := database.NewBookRepo(db)
	coverRepo := database.NewCoverRepo(db)
	pageService := NewPageService(bookRepo, nil, mockPageBuilder, mockCoverBuilder)
	thumbnailService := NewThumbnailService(pageService)

	realScanner := NewRealScanner(KnownSeries, SkipTitles)
	cfg := &Config{LibraryName: "Test Library", Host: ":0"}
	scanService := NewScanService(db, cfg, libraryRepo, seriesRepo, bookRepo, coverRepo, realScanner)

	return NewHandlers(libraryRepo, seriesRepo, bookRepo, pageService, coverRepo, thumbnailService, scanService)
}

func getTestDB(t *testing.T) *database.DB {
	t.Helper()
	testDBMutex.Lock()
	db := testDBMap[t]
	testDBMutex.Unlock()
	if db == nil {
		t.Fatal("test database not found - make sure createTestHandlers was called")
	}
	return db
}

// ID conversion helpers
func idToString(id int64) string {
	return database.Int64ToStringID(id)
}

func invalidID(t *testing.T) string {
	t.Helper()

	newGen, err := database.GetTSIDGenerator()
	if err != nil {
		t.Fatalf("failed to create TSID generator: %v", err)
	}

	newID, err := newGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate TSID: %v", err)
	}

	return database.Int64ToStringID(newID)
}

// Test fixture for repository-based tests
type testFixture struct {
	DB          *database.DB
	LibraryRepo *database.LibraryRepo
	SeriesRepo  *database.SeriesRepo
	BookRepo    *database.BookRepo
	CoverRepo   *database.CoverRepo
}

func setupTestFixture(t *testing.T) *testFixture {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return &testFixture{
		DB:          db,
		LibraryRepo: database.NewLibraryRepo(db),
		SeriesRepo:  database.NewSeriesRepo(db),
		BookRepo:    database.NewBookRepo(db),
		CoverRepo:   database.NewCoverRepo(db),
	}
}

// Test data builders
func CreateTestLibrary(t *testing.T, repo *database.LibraryRepo, name string) *database.Library {
	t.Helper()
	lib, err := repo.EnsureLibrary(context.Background(), name)
	if err != nil {
		t.Fatalf("failed to create test library: %v", err)
	}
	return lib
}

func CreateTestSeries(t *testing.T, repo *database.SeriesRepo, libraryID int64) *database.Series {
	return CreateTestSeriesWithName(t, repo, libraryID, "Test Series")
}

func CreateTestSeriesWithName(t *testing.T, repo *database.SeriesRepo, libraryID int64, name string) *database.Series {
	t.Helper()
	series := &database.Series{
		ID:        database.HashEntityID("series", name),
		LibraryID: libraryID,
		Name:      name,
	}
	if err := repo.Upsert(context.Background(), series); err != nil {
		t.Fatalf("failed to create test series: %v", err)
	}
	return series
}

func CreateTestBook(t *testing.T, repo *database.BookRepo, name string, seriesID int64) *database.Book {
	t.Helper()
	book := &database.Book{
		ID:          database.HashEntityID("book", name, "Test Publication"),
		SeriesID:    seriesID,
		Name:        name,
		Number:      1,
		Publication: "Test Publication",
		Status:      "READY",
	}
	if err := repo.Upsert(context.Background(), book); err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}
	return book
}

func CreateTestCover(t *testing.T, repo *database.CoverRepo, seriesID int64, issueNumber int) *database.Cover {
	t.Helper()
	cover := &database.Cover{
		ID:          database.HashEntityID("cover", "series", fmt.Sprintf("issue-%d", issueNumber)),
		SeriesID:    seriesID,
		IssueNumber: issueNumber,
		Publication: "2000 AD",
	}
	if err := repo.Upsert(context.Background(), cover); err != nil {
		t.Fatalf("failed to create test cover: %v", err)
	}
	return cover
}

func CreateTestEpisode(t *testing.T, repo *database.BookRepo, bookID int64, issueNumber, part int) *database.Episode {
	t.Helper()
	episode := &database.Episode{
		ID:          database.HashEntityID("episode", fmt.Sprintf("%d-%d", issueNumber, part)),
		BookID:      bookID,
		Filename:    "test.pdf",
		IssueNumber: issueNumber,
		Title:       "Test Episode",
		Part:        part,
		PageFrom:    1,
		PageTo:      10,
	}
	if err := repo.UpsertEpisodes(context.Background(), []*database.Episode{episode}); err != nil {
		t.Fatalf("failed to create test episode: %v", err)
	}
	return episode
}
