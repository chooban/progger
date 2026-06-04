package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns default when not set", func(t *testing.T) {
		t.Parallel()
		result := getEnv("TEST_KEY_NOT_SET", "default-val")
		require.Equal(t, "default-val", result)
	})

	t.Run("returns value when set", func(t *testing.T) {
		t.Setenv("TEST_KEY_SET", "custom-val")
		result := getEnv("TEST_KEY_SET", "default-val")
		require.Equal(t, "custom-val", result)
	})

	t.Run("returns default when value is empty string", func(t *testing.T) {
		t.Setenv("TEST_KEY_EMPTY", "")
		result := getEnv("TEST_KEY_EMPTY", "default-val")
		require.Equal(t, "default-val", result)
	})
}

func TestSplitDirs(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		require.Empty(t, splitDirs(""))
	})

	t.Run("single directory", func(t *testing.T) {
		t.Parallel()
		result := splitDirs("/foo/bar")
		require.Equal(t, []string{"/foo/bar"}, result)
	})

	t.Run("multiple directories", func(t *testing.T) {
		t.Parallel()
		result := splitDirs("/foo,/bar,/baz")
		require.Equal(t, []string{"/foo", "/bar", "/baz"}, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		result := splitDirs(" /foo , /bar ")
		require.Equal(t, []string{"/foo", "/bar"}, result)
	})

	t.Run("skips empty parts", func(t *testing.T) {
		t.Parallel()
		result := splitDirs("/foo,,/bar")
		require.Equal(t, []string{"/foo", "/bar"}, result)
	})

	t.Run("trailing comma", func(t *testing.T) {
		t.Parallel()
		result := splitDirs("/foo,")
		require.Equal(t, []string{"/foo"}, result)
	})
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("SCAN_DIRS", "")
	t.Setenv("HOST", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SCAN_ON_STARTUP", "")

	cfg := Load()
	require.Equal(t, "./reader.db", cfg.DatabasePath)
	require.Empty(t, cfg.ScanDirectories)
	require.Equal(t, "2000 AD", cfg.LibraryName)
	require.Equal(t, ":8081", cfg.Host)
	require.Equal(t, "info", cfg.LogLevel)
	require.False(t, cfg.ScanOnStartup)
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_PATH", "/tmp/test.db")
	t.Setenv("SCAN_DIRS", "/scans/foo,/scans/bar")
	t.Setenv("LIBRARY_NAME", "My Library")
	t.Setenv("HOST", ":9999")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SCAN_ON_STARTUP", "true")

	cfg := Load()
	require.Equal(t, "/tmp/test.db", cfg.DatabasePath)
	require.Equal(t, []string{"/scans/foo", "/scans/bar"}, cfg.ScanDirectories)
	require.Equal(t, "My Library", cfg.LibraryName)
	require.Equal(t, ":9999", cfg.Host)
	require.Equal(t, "debug", cfg.LogLevel)
	require.True(t, cfg.ScanOnStartup)
}
