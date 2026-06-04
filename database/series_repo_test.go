package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeriesRepo_Upsert(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")

	s := &Series{
		ID:        HashEntityID("series", "Dredd"),
		LibraryID: lib.ID,
		Name:      "Dredd",
	}
	err := seriesRepo.Upsert(context.Background(), s)
	require.NoError(t, err)
	require.NotEmpty(t, s.CreatedAt)
	require.NotEmpty(t, s.UpdatedAt)
}

func TestSeriesRepo_Upsert_FailsWithoutID(t *testing.T) {
	db := createTestDB(t)
	repo := NewSeriesRepo(&DB{DB: db})

	err := repo.Upsert(context.Background(), &Series{Name: "NoID"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ID must be set")
}

func TestSeriesRepo_GetByID(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	created := createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	found, err := seriesRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "Dredd", found.Name)
}

func TestSeriesRepo_MaybeGetByName(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	found, err := seriesRepo.MaybeGetByName(context.Background(), "Dredd")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "Dredd", found.Name)
}

func TestSeriesRepo_List(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	createTestSeries(t, seriesRepo, lib.ID, "B Series")
	createTestSeries(t, seriesRepo, lib.ID, "A Series")

	series, total, err := seriesRepo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Equal(t, "A Series", series[0].Name)
	require.Equal(t, "B Series", series[1].Name)
}

func TestSeriesRepo_SetBookCount(t *testing.T) {
	db := createTestDB(t)
	libRepo := NewLibraryRepo(&DB{DB: db})
	seriesRepo := NewSeriesRepo(&DB{DB: db})

	lib := createTestLibrary(t, libRepo, "Test Library")
	series := createTestSeries(t, seriesRepo, lib.ID, "Dredd")

	err := seriesRepo.SetBookCount(context.Background(), series.ID, 42)
	require.NoError(t, err)

	updated, err := seriesRepo.GetByID(context.Background(), series.ID)
	require.NoError(t, err)
	require.Equal(t, 42, updated.BookCount)
}
