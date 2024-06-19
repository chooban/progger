package api

import "fmt"

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
	Url         string
	IssueNumber int
	IssueDate   string
	Downloads   map[FileType]string
}

func (d *DigitalComic) Filename(f FileType) string {
	return fmt.Sprintf("2000AD %d (1977).%s", d.IssueNumber, f.String())
}
