package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLibraryService_EnsureLibrary_CreatesNewLibrary(t *testing.T) {
	t.Parallel()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)
	service := NewLibraryService(dbService, tsidGen)

	lib, err := service.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)
	require.NotNil(t, lib)
	require.Equal(t, "Test Library", lib.Name)
	require.Greater(t, lib.ID, int64(0))
}

func TestLibraryService_EnsureLibrary_ReturnsExistingLibrary(t *testing.T) {
	t.Parallel()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)
	service := NewLibraryService(dbService, tsidGen)

	lib1, err := service.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)

	lib2, err := service.EnsureLibrary(context.Background(), "Test Library")
	require.NoError(t, err)
	require.Equal(t, lib1.ID, lib2.ID, "should return same library ID")
	require.Equal(t, lib1.Name, lib2.Name)
}

func TestLibraryService_List_ReturnsAllLibraries(t *testing.T) {
	t.Parallel()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)
	service := NewLibraryService(dbService, tsidGen)

	_ = CreateTestLibrary(t, service, "First Library")
	_ = CreateTestLibrary(t, service, "Second Library")

	libs, err := service.List(context.Background())
	require.NoError(t, err)
	require.Len(t, libs, 2)
}

func TestLibraryService_GetByID_Found(t *testing.T) {
	t.Parallel()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)
	service := NewLibraryService(dbService, tsidGen)

	lib := CreateTestLibrary(t, service, "Test Library")

	foundLib, err := service.GetByID(context.Background(), lib.ID)
	require.NoError(t, err)
	require.NotNil(t, foundLib)
	require.Equal(t, lib.ID, foundLib.ID)
	require.Equal(t, lib.Name, foundLib.Name)
}

func TestLibraryService_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := CreateTestDB(t)
	dbService := &DB{DB: db}
	tsidGen := createTestTSIDGenerator(t)
	service := NewLibraryService(dbService, tsidGen)

	lib, err := service.GetByID(context.Background(), 999)
	require.Error(t, err)
	require.Nil(t, lib)
}
