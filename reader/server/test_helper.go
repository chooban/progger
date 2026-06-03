package server

import (
	"context"
	"image"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/chooban/progger/reader/services"
	scanApi "github.com/chooban/progger/scan/api"
	"github.com/gin-gonic/gin"
)

func mockCoverBuilder(ctx context.Context, page scanApi.ExportPage) (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, 100, 100)), nil
}

func mockPageBuilder(ctx context.Context, page scanApi.ExportPage) (*[]byte, error) {
	a := make([]byte, 100)
	return &a, nil
}

func setupTestRouter(handlers *Handlers) *gin.Engine {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	ConfigureRoutes(router, handlers)

	return router
}

var (
	testDBMutex sync.Mutex
	testDBMap   = make(map[*testing.T]*services.DB)
)

func createTestHandlers(t *testing.T) *Handlers {
	t.Helper()

	db := services.CreateTestDB(t)
	dbService := &services.DB{DB: db}
	
	testDBMutex.Lock()
	testDBMap[t] = dbService
	testDBMutex.Unlock()
	
	t.Cleanup(func() {
		testDBMutex.Lock()
		delete(testDBMap, t)
		testDBMutex.Unlock()
	})

	tsidGen, err := services.GetTSIDGenerator()
	if err != nil {
		t.Fatalf("failed to create TSID generator: %v", err)
	}

	libraryService := services.NewLibraryService(dbService, tsidGen)
	seriesService := services.NewSeriesService(dbService, tsidGen)
	bookService := services.NewBookService(dbService, tsidGen)
	coverService := services.NewCoverService(dbService, tsidGen)
	pageService := services.NewPageService(bookService, nil, mockPageBuilder, mockCoverBuilder)
	thumbnailService := services.NewThumbnailService(pageService)
	scanService := services.NewScanService(dbService, nil, libraryService, seriesService, bookService, coverService, nil)

	return NewHandlers(libraryService, seriesService, bookService, pageService, coverService, thumbnailService, scanService)
}

func createTestServer(t *testing.T, handlers *Handlers) *httptest.Server {
	t.Helper()

	router := setupTestRouter(handlers)
	return httptest.NewServer(router)
}

// idToString converts an int64 TSID to its 13-character base32 string representation
func idToString(id int64) string {
	return services.Int64ToStringID(id)
}

// invalidID returns a TSID string that doesn't exist in the database
// We generate a fresh TSID from a new generator instance
func invalidID(t *testing.T) string {
	t.Helper()

	newGen, err := services.GetTSIDGenerator()
	if err != nil {
		t.Fatalf("failed to create TSID generator: %v", err)
	}

	newID, err := newGen.Generate()
	if err != nil {
		t.Fatalf("failed to generate TSID: %v", err)
	}

	return services.Int64ToStringID(newID)
}

// getTestDB returns the test database instance for the current test
func getTestDB(t *testing.T) *services.DB {
	t.Helper()
	testDBMutex.Lock()
	db := testDBMap[t]
	testDBMutex.Unlock()
	if db == nil {
		t.Fatal("test database not found - make sure createTestHandlers was called")
	}
	return db
}
