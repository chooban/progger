package pdf

import (
	"github.com/chooban/progdl-go/internal"
	"github.com/chooban/progdl-go/internal/db"
)

type Reader interface {
	Bookmarks(filename string) ([]internal.EpisodeDetails, error)
	Build(episodes []db.Episode)
	Credits(filename string, startPage int, endPage int) (string, error)
}
