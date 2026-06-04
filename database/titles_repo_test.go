package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKnownTitlesRepo_ListEmpty(t *testing.T) {
	db := createTestDB(t)
	repo := NewKnownTitlesRepo(&DB{DB: db})

	titles, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, titles)
}

func TestKnownTitlesRepo_AddAndList(t *testing.T) {
	db := createTestDB(t)
	repo := NewKnownTitlesRepo(&DB{DB: db})

	require.NoError(t, repo.Add(context.Background(), "Dredd"))
	require.NoError(t, repo.Add(context.Background(), "Strontium Dog"))

	titles, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Contains(t, titles, "Dredd")
	require.Contains(t, titles, "Strontium Dog")
}

func TestKnownTitlesRepo_Delete(t *testing.T) {
	db := createTestDB(t)
	repo := NewKnownTitlesRepo(&DB{DB: db})

	require.NoError(t, repo.Add(context.Background(), "Dredd"))
	require.NoError(t, repo.Delete(context.Background(), "Dredd"))

	titles, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, titles)
}

func TestSkipTitlesRepo_AddAndList(t *testing.T) {
	db := createTestDB(t)
	repo := NewSkipTitlesRepo(&DB{DB: db})

	require.NoError(t, repo.Add(context.Background(), "Untitled"))
	require.NoError(t, repo.Add(context.Background(), "Obituary"))

	titles, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Contains(t, titles, "Untitled")
	require.Contains(t, titles, "Obituary")
}

func TestSkipTitlesRepo_Delete(t *testing.T) {
	db := createTestDB(t)
	repo := NewSkipTitlesRepo(&DB{DB: db})

	require.NoError(t, repo.Add(context.Background(), "Untitled"))
	require.NoError(t, repo.Delete(context.Background(), "Untitled"))

	titles, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, titles)
}
