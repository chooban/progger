package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageRequest_Offset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  PageRequest
		want int
	}{
		{"page zero", PageRequest{Page: 0, Size: 10}, 0},
		{"page one", PageRequest{Page: 1, Size: 10}, 0},
		{"page two", PageRequest{Page: 2, Size: 10}, 10},
		{"negative page", PageRequest{Page: -1, Size: 10}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.req.Offset())
		})
	}
}

func TestPageRequest_Limit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  PageRequest
		want int
	}{
		{"normal", PageRequest{Page: 1, Size: 10}, 10},
		{"unpaged", PageRequest{Page: 1, Size: 10, Unpaged: true}, -1},
		{"zero size", PageRequest{Page: 1, Size: 0}, -1},
		{"negative size", PageRequest{Page: 1, Size: -5}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.req.Limit())
		})
	}
}

func TestNewPageResponse(t *testing.T) {
	t.Parallel()

	t.Run("empty book slices", func(t *testing.T) {
		t.Parallel()
		resp := NewPageResponse([]BookDto{}, 0, 1, 10)
		require.True(t, resp.Empty)
		require.Equal(t, int32(0), resp.NumberOfElements)
		require.True(t, resp.Last)
	})

	t.Run("non-empty book slice", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}, {ID: "2"}, {ID: "3"}}
		resp := NewPageResponse(books, 3, 1, 10)
		require.False(t, resp.Empty)
		require.Equal(t, int32(3), resp.NumberOfElements)
		require.Equal(t, int32(3), resp.NumberOfElements)
		require.Equal(t, int64(3), resp.TotalElements)
	})

	t.Run("series slice", func(t *testing.T) {
		t.Parallel()
		series := []SeriesDto{{ID: "a"}}
		resp := NewPageResponse(series, 1, 1, 5)
		require.False(t, resp.Empty)
		require.Equal(t, int32(1), resp.NumberOfElements)
	})

	t.Run("collection slice", func(t *testing.T) {
		t.Parallel()
		cols := []CollectionDto{{ID: "c"}}
		resp := NewPageResponse(cols, 1, 1, 5)
		require.False(t, resp.Empty)
		require.Equal(t, int32(1), resp.NumberOfElements)
	})

	t.Run("readlist slice", func(t *testing.T) {
		t.Parallel()
		rls := []ReadListDto{{ID: "r"}}
		resp := NewPageResponse(rls, 1, 1, 5)
		require.False(t, resp.Empty)
		require.Equal(t, int32(1), resp.NumberOfElements)
	})

	t.Run("totalPages with exact division", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 10, 1, 5)
		require.Equal(t, int32(2), resp.TotalPages)
	})

	t.Run("totalPages with remainder", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 10, 1, 3)
		require.Equal(t, int32(4), resp.TotalPages)
	})

	t.Run("totalPages zero size", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 10, 1, 0)
		require.Equal(t, int32(0), resp.TotalPages)
		require.Equal(t, int32(0), resp.Size)
	})

	t.Run("totalPages minimum one", func(t *testing.T) {
		t.Parallel()
		resp := NewPageResponse([]BookDto{}, 0, 1, 5)
		require.Equal(t, int32(1), resp.TotalPages)
	})

	t.Run("first page flag", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 5, 1, 5)
		require.True(t, resp.First)
	})

	t.Run("not first page flag", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 5, 2, 5)
		require.False(t, resp.First)
	})

	t.Run("last page flag", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}}
		resp := NewPageResponse(books, 5, 1, 5)
		require.True(t, resp.Last)
	})

	t.Run("not last page flag", func(t *testing.T) {
		t.Parallel()
		books := []BookDto{{ID: "1"}, {ID: "2"}}
		resp := NewPageResponse(books, 10, 1, 2)
		require.False(t, resp.Last)
	})

	t.Run("last page when count zero", func(t *testing.T) {
		t.Parallel()
		resp := NewPageResponse([]BookDto{}, 10, 1, 5)
		require.True(t, resp.Last)
	})
}

func TestFormatInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"positive", 42, "42"},
		{"zero", 0, "0"},
		{"negative", -7, "-7"},
		{"large", 1234567890, "1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, FormatInt(tt.n))
		})
	}
}
