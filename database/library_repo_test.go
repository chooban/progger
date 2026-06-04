package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLibraryRepo_EnsureLibrary_Creates(t *testing.T) {
	db := createTestDB(t)
	repo := NewLibraryRepo(&DB{DB: db})

	lib, err := repo.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)
	require.Equal(t, "Test Library", lib.Name)
	require.NotZero(t, lib.ID)
	require.NotEmpty(t, lib.CreatedAt)
	require.NotEmpty(t, lib.UpdatedAt)
}

func TestLibraryRepo_EnsureLibrary_Idempotent(t *testing.T) {
	db := createTestDB(t)
	repo := NewLibraryRepo(&DB{DB: db})

	lib1, err := repo.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)

	lib2, err := repo.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)
	require.Equal(t, lib1.ID, lib2.ID)
}

func TestLibraryRepo_List(t *testing.T) {
	db := createTestDB(t)
	repo := NewLibraryRepo(&DB{DB: db})

	_, _ = repo.EnsureLibrary(context.Background(), "B Library")
	_, _ = repo.EnsureLibrary(context.Background(), "A Library")

	libs, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, libs, 2)
	require.Equal(t, "A Library", libs[0].Name)
	require.Equal(t, "B Library", libs[1].Name)
}

func TestLibraryRepo_GetByID(t *testing.T) {
	db := createTestDB(t)
	repo := NewLibraryRepo(&DB{DB: db})

	created, _ := repo.EnsureLibrary(context.Background(), "Test Library")
	found, err := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.Name, found.Name)
}
