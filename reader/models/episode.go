package models

type Episode struct {
	ID          int64   `db:"id"`
	BookID      int64   `db:"book_id"`
	Filename    string  `db:"filename"`
	IssueNumber int     `db:"issue_number"`
	Title       string  `db:"title"`
	Part        int     `db:"part"`
	PageFrom    int     `db:"page_from"`
	PageTo      int     `db:"page_to"`
	ReleaseDate *string `db:"release_date"`
}
