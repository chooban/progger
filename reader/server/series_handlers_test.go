package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chooban/progger/reader/api"
	"github.com/chooban/progger/reader/services"
	"github.com/stretchr/testify/require"
)

func TestListSeries_ReturnsPaginatedSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	_ = services.CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = services.CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	reqBody := `{"page":0,"size":10}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/series/list", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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

func TestGetSeries_Found(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	series := services.CreateTestSeries(t, handlers.seriesSer, lib.ID)

	resp, err := http.Get(server.URL + "/api/v1/series/" + idToString(series.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var s api.SeriesDto
	json.NewDecoder(resp.Body).Decode(&s)
	require.Equal(t, "Test Series", s.Name)
}

func TestGetSeries_NotFound(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/series/" + invalidID(t))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListSeriesThumbnails_ReturnsList(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	ser := services.CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")
	cover := services.CreateTestCover(t, handlers.coverSer, ser.ID, 100)
	require.NotEqual(t, 0, cover.ID)

	resp, err := http.Get(server.URL + "/api/v1/series/" + idToString(ser.ID) + "/thumbnails")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var thumbs []api.ThumbnailDto
	json.NewDecoder(resp.Body).Decode(&thumbs)
	require.Len(t, thumbs, 1)
	require.Equal(t, "series", thumbs[0].Type)
	require.NotEmpty(t, thumbs[0].ID)
	// ID should be a TSID string format (13 characters)
	require.Len(t, thumbs[0].ID, 13)
}

func TestRecentlyAddedSeries_ReturnsSeriesAddedInLast7Days(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	_ = services.CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = services.CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	resp, err := http.Get(server.URL + "/api/v1/series/new")
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

func TestRecentlyAddedSeries_EmptyWhenNoNewSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/series/new")
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

func TestRecentlyUpdatedSeries_ReturnsSeriesUpdatedInLast7Days(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	_ = services.CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = services.CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	resp, err := http.Get(server.URL + "/api/v1/series/updated")
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

func TestRecentlyUpdatedSeries_EmptyWhenNoUpdatedSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/series/updated")
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

func TestListSeriesLatest_ReturnsPaginatedSeries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")
	_ = services.CreateTestSeries(t, handlers.seriesSer, lib.ID)
	_ = services.CreateTestSeriesWithName(t, handlers.seriesSer, lib.ID, "Series B")

	resp, err := http.Get(server.URL + "/api/v1/series/latest")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	var pageResp2 map[string]interface{}
	json.Unmarshal(body, &pageResp2)
	if content, ok := pageResp2["content"].([]interface{}); ok {
		require.Len(t, content, 2)
	}
	require.Equal(t, float64(2), pageResp2["totalElements"])
}
