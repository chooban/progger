package database

import (
	"github.com/rushysloth/go-tsid"
)

type Series struct {
	ID        int64  `db:"id"`
	LibraryID int64  `db:"library_id"`
	Name      string `db:"name"`
	BookCount int    `db:"book_count"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (s Series) URL() string {
	t := tsid.FromNumber(s.ID)
	return "/api/v1/series/" + t.ToString()
}
