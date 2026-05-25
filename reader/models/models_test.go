package models

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/rushysloth/go-tsid"
	"github.com/stretchr/testify/require"
)

func newTSID(t *testing.T) int64 {
	t.Helper()
	factory, err := tsid.TsidFactoryBuilder().WithNodeBits(0).Build()
	require.NoError(t, err)
	tsidVal, err := factory.Generate()
	require.NoError(t, err)
	return tsidVal.ToNumber()
}

func TestBook_URL(t *testing.T) {
	t.Parallel()

	id := newTSID(t)
	book := Book{ID: id}
	expected := fmt.Sprintf("/api/v1/books/%s", tsid.FromNumber(id).ToString())
	require.Equal(t, expected, book.URL())
}

func TestSeries_URL(t *testing.T) {
	t.Parallel()

	id := newTSID(t)
	series := Series{ID: id}
	expected := fmt.Sprintf("/api/v1/series/%s", tsid.FromNumber(id).ToString())
	require.Equal(t, expected, series.URL())
}

func TestLibrary_URL(t *testing.T) {
	t.Parallel()

	id := newTSID(t)
	library := Library{ID: id}
	expected := fmt.Sprintf("/api/v1/libraries/%s", tsid.FromNumber(id).ToString())
	require.Equal(t, expected, library.URL())
}

func TestCover_URL(t *testing.T) {
	t.Parallel()

	cover := Cover{ID: 42}
	expected := "/api/v1/covers/" + strconv.FormatInt(42, 10)
	require.Equal(t, expected, cover.URL())
}

func TestBuildPage(t *testing.T) {
	t.Parallel()

	page := BuildPage(100, 5)
	require.Equal(t, 5, page.Number)
	require.Equal(t, "application/pdf", page.MediaType)
	require.Equal(t, "100-page-5.pdf", page.FileName)
	require.Equal(t, 1630, page.Height)
	require.Equal(t, 1241, page.Width)
	require.Empty(t, page.Size)
	require.Equal(t, int64(0), page.SizeBytes)
}
