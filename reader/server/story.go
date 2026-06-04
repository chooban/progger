package server

import scanApi "github.com/chooban/progger/scan/api"

type Story struct {
	Title      string
	Series     string
	FirstIssue int
	LastIssue  int
	Episodes   []Episode
}

type Episode struct {
	*scanApi.Episode
	Filename    string
	IssueNumber int
}
