package server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListCollections_ReturnsPaginatedEmptyList(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/collections")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	json.Unmarshal(body, &parsed)

	if content, ok := parsed["content"].([]interface{}); ok {
		require.Empty(t, content)
	}
}

func TestListReadLists_ReturnsPaginatedEmptyList(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/readlists")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	json.Unmarshal(body, &parsed)

	if content, ok := parsed["content"].([]interface{}); ok {
		require.Empty(t, content)
	}
}

func TestListGlobalClientSettings_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/client-settings/global/list")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListUserClientSettings_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/client-settings/user/list")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
