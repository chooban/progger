package api

import (
	"fmt"
	"time"
)

type FileType int

const (
	Pdf FileType = iota
	Cbz
)

func (f *FileType) String() string {
	names := [...]string{"pdf", "cbz"}
	return names[*f]
}

type DigitalComic struct {
	Publication string
	Url         string
	IssueNumber int
	IssueDate   string
	Downloads   map[FileType]string
}

func (d *DigitalComic) Filename(f FileType) string {
	return fmt.Sprintf("%s %d (1977).%s", d.Publication, d.IssueNumber, f.String())
}

func (d *DigitalComic) String() string {
	issueDate, _ := time.Parse("2006-01-02", d.IssueDate)
	formattedDate := formatDateWithOrdinal(issueDate)

	return fmt.Sprintf("%s %d (%s)", d.Publication, d.IssueNumber, formattedDate)
}

// formatDateWithOrdinal prints a given time in the format 1st January 2000.
func formatDateWithOrdinal(t time.Time) string {
	return fmt.Sprintf("%s %s %d", addOrdinal(t.Day()), t.Month(), t.Year())
}

// addOrdinal takes a number and adds its ordinal (like st or th) to the end.
func addOrdinal(n int) string {
	switch n {
	case 1, 21, 31:
		return fmt.Sprintf("%dst", n)
	case 2, 22:
		return fmt.Sprintf("%dnd", n)
	case 3, 23:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
