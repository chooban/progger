package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chooban/progger/database"
	"github.com/stretchr/testify/require"
)

func TestListBooks_EmptyBody_ReturnsAllBooks(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book1 := CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Book", series1.ID)

	reqBody := `{}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 2)
		ids := map[int64]bool{}
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if idStr, ok := itemMap["id"].(string); ok {
					if id, err := database.StringIDToInt64(idStr); err == nil {
						ids[id] = true
					}
				}
			}
		}
		require.Len(t, ids, 2)
		require.True(t, ids[book1.ID])
		require.True(t, ids[book2.ID])
	}
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooks_WithSeriesId_FiltersCorrectly(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	book1 := CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Second", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Third", series2.ID)

	series1ID := idToString(series1.ID)
	reqBody := `{"condition":{"seriesId":{"value":"` + series1ID + `"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	require.Equal(t, float64(2), pageResp["totalElements"])

	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 2)
		ids := map[int64]bool{}
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if idStr, ok := itemMap["id"].(string); ok {
					if id, err := database.StringIDToInt64(idStr); err == nil {
						ids[id] = true
					}
				}
			}
		}
		require.Len(t, ids, 2)
		require.True(t, ids[book1.ID])
		require.True(t, ids[book2.ID])
	}
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooks_WithLibraryId_FiltersCorrectly(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib1 := CreateTestLibrary(t, handlers.librarySer, "Library A")
	lib2 := CreateTestLibrary(t, handlers.librarySer, "Library B")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib1.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib2.ID, "Test Series 2")

	_ = CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Second", series2.ID)

	lib1ID := idToString(lib1.ID)
	reqBody := `{"condition":{"libraryId":{"value":"` + lib1ID + `"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	require.Equal(t, float64(1), pageResp["totalElements"])

	if content, ok := pageResp["content"].([]interface{}); ok && len(content) > 0 {
		if book, ok := content[0].(map[string]interface{}); ok {
			if seriesID, ok := book["seriesId"].(string); ok {
				seriesIDs := idToString(series1.ID)
				require.Equal(t, seriesIDs, seriesID)
			}
		}
	}
}

func TestListBooks_InvalidSeriesId_Returns400(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	reqBody := `{"condition":{"seriesId":{"value":"invalid"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListBooks_InvalidLibraryId_Returns400(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	reqBody := `{"condition":{"libraryId":{"value":"invalid"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetBook_Found(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book := CreateTestBook(t, handlers.bookSer, "First", series.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var b BookDto
	json.NewDecoder(resp.Body).Decode(&b)
	require.Equal(t, "1 First", b.Name)
}

func TestGetBook_NotFound(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidID(t))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListBooksOnDeck_ReturnsFirstBookPerSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	book1 := CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Second", series1.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Third", series2.ID)

	// Ensure book1 has the lowest ID in series1 for the MIN(id) OnDeck query
	_, err := getTestDB(t).ExecContext(context.Background(), "UPDATE books SET id = 1 WHERE id = ?", book1.ID)
	require.NoError(t, err)
	book1.ID = 1

	resp, err := http.Get(server.URL + "/api/v1/books/ondeck")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 2)
		ids := map[int64]bool{}
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if idStr, ok := itemMap["id"].(string); ok {
					if id, err := database.StringIDToInt64(idStr); err == nil {
						ids[id] = true
					}
				}
			}
		}
		require.Len(t, ids, 2)
		require.True(t, ids[book1.ID])
		require.True(t, ids[book2.ID])
	}
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooksOnDeck_WithLibraryId_FiltersCorrectly(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib1 := CreateTestLibrary(t, handlers.librarySer, "Library A")
	lib2 := CreateTestLibrary(t, handlers.librarySer, "Library B")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib1.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib2.ID, "Series B")

	book1 := CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Second", series2.ID)

	lib1ID := idToString(lib1.ID)
	resp, err := http.Get(server.URL + "/api/v1/books/ondeck?library_id=" + lib1ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok && len(content) > 0 {
		if book, ok := content[0].(map[string]interface{}); ok {
			if idStr, ok := book["id"].(string); ok {
				id, _ := database.StringIDToInt64(idStr)
				require.Equal(t, book1.ID, id)
			}
		}
	}
	require.Equal(t, float64(1), pageResp["totalElements"])
}

func TestListBooksOnDeck_EmptyWhenNoBooks(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/books/ondeck")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Empty(t, content)
	}
	require.Equal(t, float64(0), pageResp["totalElements"])
}

func TestListBookThumbnails_ReturnsThumbnailMetadata(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book := CreateTestBook(t, handlers.bookSer, "Test Book", series.ID)
	_ = CreateTestEpisode(t, handlers.bookSer, book.ID, 1, 1)
	_ = CreateTestCover(t, handlers.coverSer, series.ID, 1)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book.ID) + "/thumbnails")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var thumbs []ThumbnailDto
	json.NewDecoder(resp.Body).Decode(&thumbs)
	require.Len(t, thumbs, 1)
	require.Equal(t, "book", thumbs[0].Type)
}

func TestListBooksV1_ReturnsBooksBySeriesID(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Book 1", series.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Book 2", series.ID)

	// List books for the series
	resp, err := http.Get(server.URL + "/api/v1/series/" + idToString(series.ID) + "/books")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 2)
	}
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooksV1_ReturnsBadRequestForInvalidSeriesID(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Try with invalid series ID
	resp, err := http.Get(server.URL + "/api/v1/series/invalid/books")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListBooksV1_Returns404ForNonExistentSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	invalidSeriesID := database.Int64ToStringID(1000000000)

	// Try with non-existent series
	resp, err := http.Get(server.URL + "/api/v1/series/" + invalidSeriesID + "/books")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListBooksLatest_ReturnsPaginatedBooks(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = CreateTestBook(t, handlers.bookSer, "First", series.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Second", series.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/latest")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 2)
	}
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooksLatest_FiltersByLibrary(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib1 := CreateTestLibrary(t, handlers.librarySer, "Library 1")
	lib2 := CreateTestLibrary(t, handlers.librarySer, "Library 2")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib1.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib2.ID, "Series 2")

	_ = CreateTestBook(t, handlers.bookSer, "Book 1", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Book 2", series2.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/latest?library_id=" + idToString(lib1.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	if content, ok := pageResp["content"].([]interface{}); ok {
		require.Len(t, content, 1)
	}
	require.Equal(t, float64(1), pageResp["totalElements"])
}

func TestListBooksLatest_InvalidLibraryId_Returns400(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/books/latest?library_id=invalid")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetBookSiblingPrevious_NotFound(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = CreateTestBook(t, handlers.bookSer, "First", series.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Second", series.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book2.ID) + "/previous")
	require.NoError(t, err)
	// Both books have number=1 so no previous exists
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetBookSiblingPrevious_Found(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book1 := CreateTestBook(t, handlers.bookSer, "First", series.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Second", series.ID)

	// Set sequential numbers for ordering
	_, err := getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", book1.ID)
	require.NoError(t, err)
	_, err = getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", book2.ID)
	require.NoError(t, err)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book2.ID) + "/previous")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var book BookDto
	json.NewDecoder(resp.Body).Decode(&book)
	require.Equal(t, "1 First", book.Name)
	require.Equal(t, idToString(book1.ID), book.ID)
}

func TestGetBookSiblingNext_Found(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book1 := CreateTestBook(t, handlers.bookSer, "First", series.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Second", series.ID)

	// Set sequential numbers for ordering
	_, err := getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", book1.ID)
	require.NoError(t, err)
	_, err = getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", book2.ID)
	require.NoError(t, err)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book1.ID) + "/next")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var book BookDto
	json.NewDecoder(resp.Body).Decode(&book)
	require.Equal(t, "2 Second", book.Name)
	require.Equal(t, idToString(book2.ID), book.ID)
}

func TestGetBookSiblingNext_NotFound(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	last := CreateTestBook(t, handlers.bookSer, "Last", series.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(last.ID) + "/next")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDownloadBook_Returns404ForMissingBook(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidID(t) + "/file")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDownloadBook_Returns404ForBookWithNoEpisodes(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book := CreateTestBook(t, handlers.bookSer, "No Episodes", series.ID)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book.ID) + "/file")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetBookSiblingNext_Returns404WhenSeriesDeleted(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book1 := CreateTestBook(t, handlers.bookSer, "First", series.ID)
	book2 := CreateTestBook(t, handlers.bookSer, "Second", series.ID)

	// Set sequential numbers
	_, err := getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 1 WHERE id = ?", book1.ID)
	require.NoError(t, err)
	_, err = getTestDB(t).ExecContext(context.Background(), "UPDATE books SET number = 2 WHERE id = ?", book2.ID)
	require.NoError(t, err)

	// Delete the books first to satisfy FK constraints, then the series
	_, err = getTestDB(t).ExecContext(context.Background(), "DELETE FROM books WHERE series_id = ?", series.ID)
	require.NoError(t, err)
	_, err = getTestDB(t).ExecContext(context.Background(), "DELETE FROM series WHERE id = ?", series.ID)
	require.NoError(t, err)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book1.ID) + "/next")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListBooks_WithIsOperator_FiltersCorrectly(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series1 := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	series2 := CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	_ = CreateTestBook(t, handlers.bookSer, "First", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Second", series1.ID)
	_ = CreateTestBook(t, handlers.bookSer, "Third", series2.ID)

	series1ID := idToString(series1.ID)
	reqBody := `{"condition":{"seriesId":{"operator":"is","value":"` + series1ID + `"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var pageResp map[string]interface{}
	json.Unmarshal(body, &pageResp)
	require.Equal(t, float64(2), pageResp["totalElements"])
}

func TestListBooks_InvalidOperator_Returns400(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)

	seriesID := idToString(series.ID)
	reqBody := `{"condition":{"seriesId":{"operator":"invalid","value":"` + seriesID + `"}}}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListBooks_MalformedJSON_Returns400(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/books/list", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetBookThumbnail_NoCovers_FallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book := CreateTestBook(t, handlers.bookSer, "Test Book", series.ID)
	_ = CreateTestEpisode(t, handlers.bookSer, book.ID, 1, 1)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book.ID) + "/thumbnail")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	require.Greater(t, len(body), 0)
}

func TestGetBookThumbnail_WithCovers_ReturnsCover(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := CreateTestSeries(t, handlers.seriesSer, lib.ID)
	book := CreateTestBook(t, handlers.bookSer, "Test Book", series.ID)
	_ = CreateTestEpisode(t, handlers.bookSer, book.ID, 1, 1)
	_ = CreateTestCover(t, handlers.coverSer, series.ID, 1)

	resp, err := http.Get(server.URL + "/api/v1/books/" + idToString(book.ID) + "/thumbnail")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	require.Greater(t, len(body), 0)
}
