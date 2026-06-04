package server

import "strconv"

type Page struct {
	Number    int    `json:"number"`
	MediaType string `json:"mediaType"`
	FileName  string `json:"fileName"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"sizeBytes"`
}

func BuildPage(bookID int64, num int) Page {
	return Page{
		Number:    num,
		MediaType: "application/pdf",
		FileName:  strconv.FormatInt(bookID, 10) + "-page-" + strconv.FormatInt(int64(num), 10) + ".pdf",
		Height:    1630,
		Width:     1241,
		Size:      "",
		SizeBytes: 0,
	}
}
