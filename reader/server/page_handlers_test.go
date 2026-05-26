package server

import (
	"net/http"
	"testing"

	"github.com/chooban/progger/reader/services"
	"github.com/stretchr/testify/require"
)

func TestListPages_Returns404ForNonExistentBook(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an int64 that's guaranteed to not exist as a TSID
	invalidBookID := services.Int64ToStringID(1000000000)

	// Use the invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetPageThumbnail_Returns404ForNonExistentBook(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an int64 that's guaranteed to not exist as a TSID
	invalidBookID := services.Int64ToStringID(1000000000)

	// Use the invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/1/thumbnail")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetPageThumbnail_ReturnsBadRequestForInvalidPageNumber(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an int64 that's guaranteed to not exist as a TSID
	invalidBookID := services.Int64ToStringID(1000000000)

	// Use the invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/invalid/thumbnail")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPageThumbnail_ReturnsBadRequestForZeroPageNumber(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an int64 that's guaranteed to not exist as a TSID
	invalidBookID := services.Int64ToStringID(1000000000)

	// Use the invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/0/thumbnail")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPage_ReturnsBadRequestForInvalidBookID(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use invalid book ID
	resp, err := http.Get(server.URL + "/api/v1/books/invalid/pages/1")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPage_ReturnsBadRequestForInvalidPageNumber(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use a valid book ID but invalid page number
	invalidBookID := services.Int64ToStringID(1000000000)

	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/invalid")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPage_ReturnsBadRequestForZeroPageNumber(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use a valid book ID but page number 0
	invalidBookID := services.Int64ToStringID(1000000000)

	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/0")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPage_Returns404ForNonExistentBook(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use a valid format TSID but non-existent book
	invalidBookID := services.Int64ToStringID(1000000000)

	resp, err := http.Get(server.URL + "/api/v1/books/" + invalidBookID + "/pages/1")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListPages_ReturnsBadRequestForInvalidBookID(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use invalid book ID
	resp, err := http.Get(server.URL + "/api/v1/books/invalid/pages")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
