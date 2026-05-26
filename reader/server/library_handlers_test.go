package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/chooban/progger/reader/api"
	"github.com/chooban/progger/reader/services"
	"github.com/stretchr/testify/require"
)

func TestListLibraries_ReturnsAllLibraries(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib1 := services.CreateTestLibrary(t, handlers.librarySer, "Library A")
	lib2 := services.CreateTestLibrary(t, handlers.librarySer, "Library B")

	resp, err := http.Get(server.URL + "/api/v1/libraries")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var libs []api.LibraryDto
	json.NewDecoder(resp.Body).Decode(&libs)
	require.Len(t, libs, 2)

	ids := map[string]bool{}
	for _, lib := range libs {
		ids[lib.ID] = true
	}
	require.True(t, ids[idToString(lib1.ID)])
	require.True(t, ids[idToString(lib2.ID)])
}

func TestGetLibrary_Found(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	lib := services.CreateTestLibrary(t, handlers.librarySer, "Test Library")

	resp, err := http.Get(server.URL + "/api/v1/libraries/" + idToString(lib.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var library api.LibraryDto
	json.NewDecoder(resp.Body).Decode(&library)
	require.Equal(t, "Test Library", library.Name)
}

func TestGetLibrary_NotFound(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	// Use an invalid TSID string
	resp, err := http.Get(server.URL + "/api/v1/libraries/" + invalidID(t))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestScanLibrary_TriggersScan(t *testing.T) {
	t.Skip()
	tmpDir := t.TempDir()
	t.Setenv("SCAN_DIRS", tmpDir)

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/v1/libraries/1/scan", "application/json", strings.NewReader(""))
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
}
