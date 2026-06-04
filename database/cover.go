package database

import "strconv"

type Cover struct {
	ID          int64  `db:"id"`
	Text        string `db:"text"`
	SeriesID    int64  `db:"series_id"`
	Artist      string `db:"artist"`
	Filename    string `db:"filename"`
	IssueNumber int    `db:"issue_number"`
	CreatedAt   string `db:"created_at"`
	Publication string `db:"publication"`
}

func (c Cover) URL() string {
	return "/api/v1/covers/" + strconv.FormatInt(c.ID, 10)
}
