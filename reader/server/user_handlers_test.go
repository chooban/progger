package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/chooban/progger/reader/api"
	"github.com/stretchr/testify/require"
)

func TestGetMe_ReturnsUserDto(t *testing.T) {
	t.Parallel()

	handlers := createTestHandlers(t)
	server := createTestServer(t, handlers)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v2/users/me")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var user api.UserDto
	json.NewDecoder(resp.Body).Decode(&user)
	require.Equal(t, "admin@example.com", user.Email)
	require.Equal(t, "1", user.ID)
	require.Len(t, user.Roles, 1)
	require.Equal(t, "ADMIN", user.Roles[0])
}
