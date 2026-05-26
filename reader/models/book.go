package models

import (
	"github.com/rushysloth/go-tsid"
)

type Book struct {
	ID          int64      `db:"id"`
	SeriesID    int64      `db:"series_id"`
	Name        string     `db:"name"`
	Number      int        `db:"number"`
	Publication string     `db:"publication"`
	SizeBytes   int64      `db:"size_bytes"`
	FileHash    string     `db:"file_hash"`
	Status      string     `db:"status"`
	PageCount   int        `db:"page_count"`
	FirstIssue  int        `db:"first_issue"`
	LastIssue   int        `db:"last_issue"`
	ReleaseDate *string    `db:"release_date"`
	CreatedAt   string     `db:"created_at"`
	UpdatedAt   string     `db:"updated_at"`
	Episodes    []*Episode `db:"episode"`
}

func (b Book) URL() string {
	t := tsid.FromNumber(b.ID)
	return "/api/v1/books/" + t.ToString()
}

type BookWithSeries struct {
	Book       *Book
	SeriesID   int64
	SeriesName string
	Filename   string
}
