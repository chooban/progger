package server

import (
	"testing"

	"github.com/chooban/progger/reader/models"
	"github.com/stretchr/testify/require"
)

func TestComputeTotalPages(t *testing.T) {
	t.Parallel()

	t.Run("single episode", func(t *testing.T) {
		t.Parallel()
		episodes := []*models.Episode{
			{PageFrom: 1, PageTo: 10},
		}
		require.Equal(t, 10, computeTotalPages(episodes))
	})

	t.Run("multiple episodes", func(t *testing.T) {
		t.Parallel()
		episodes := []*models.Episode{
			{PageFrom: 1, PageTo: 10},
			{PageFrom: 11, PageTo: 20},
		}
		require.Equal(t, 20, computeTotalPages(episodes))
	})

	t.Run("single page episode", func(t *testing.T) {
		t.Parallel()
		episodes := []*models.Episode{
			{PageFrom: 5, PageTo: 5},
		}
		require.Equal(t, 1, computeTotalPages(episodes))
	})

	t.Run("empty episodes", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 0, computeTotalPages(nil))
		require.Equal(t, 0, computeTotalPages([]*models.Episode{}))
	})
}
