package database

import "github.com/rushysloth/go-tsid"

type Library struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (l Library) URL() string {
	t := tsid.FromNumber(l.ID)
	return "/api/v1/libraries/" + t.ToString()
}
