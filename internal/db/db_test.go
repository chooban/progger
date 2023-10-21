package db

import (
	"github.com/chooban/progdl-go/internal/env"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func createAppEnv() env.AppEnv {
	writer := zerolog.ConsoleWriter{
		Out:        io.Discard,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	appEnv := env.AppEnv{
		Log: &logger,
		Db:  Init("file::memory:?cache=shared"),
	}
	return appEnv
}

func TestSaveIssue(t *testing.T) {
	appEnv := createAppEnv()

	issue := Issue{
		Publication: Publication{Title: "Test Publication"},
		IssueNumber: 123,
		Episodes: []Episode{
			{
				Title:  "Test Episode",
				Part:   1,
				Series: Series{Title: "Test Series"},
			},
		},
	}
	SaveIssues(appEnv, []Issue{issue})

	var count int64
	appEnv.Db.Model(Publication{}).Count(&count)
	assert.Equal(t, int64(1), count)

	appEnv.Db.Model(Issue{}).Count(&count)
	assert.Equal(t, int64(1), count)

	appEnv.Db.Model(Series{}).Count(&count)
	assert.Equal(t, int64(1), count)

	appEnv.Db.Model(Episode{}).Count(&count)
	assert.Equal(t, int64(1), count)
}
