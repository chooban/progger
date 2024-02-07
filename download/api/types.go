package api

type FileType int

const (
	Pdf FileType = iota
	Cbz
)

type DigitalComic struct {
	Url         string
	IssueNumber int
	Downloads   map[FileType]string
}
