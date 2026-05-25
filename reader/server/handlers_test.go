package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthCheck_ReturnsOK(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	serverURL := server.URL

	for _, path := range []string{"", "/"} {
		resp, err := http.Get(serverURL + path)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var health HealthResponse
		json.NewDecoder(resp.Body).Decode(&health)
		require.Equal(t, "ok", health.Status)
		require.Equal(t, "0.0.1", health.Version)
		resp.Body.Close()
	}
}
