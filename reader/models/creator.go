package models

type Creator struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}
