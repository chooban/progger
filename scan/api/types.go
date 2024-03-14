package api

import (
	"errors"
	"strings"
)

// An Episode represents information extracted from a PDF bookmark.
type Episode struct {
	Series    string
	Title     string
	Part      int
	FirstPage int
	LastPage  int
	Credits   Credits
}

type Issue struct {
	Publication string
	IssueNumber int
	Episodes    []*Episode
	Filename    string
}

type Creator struct {
	Name string
}

type Credits = map[Role][]string

const (
	Unknown Role = iota
	Script
	Art
	Colours
	Letters
)

type Role int64

func NewRole(s string) (Role, error) {
	switch strings.ToLower(s) {
	case "script":
		return Script, nil
	case "art":
		return Art, nil
	case "colours":
		return Colours, nil
	case "letters":
		return Letters, nil
	}
	return Unknown, errors.New("role not found")
}

func (r Role) String() string {
	switch r {
	case Unknown:
		return "unknown"
	case Script:
		return "script"
	case Art:
		return "art"
	case Colours:
		return "colours"
	case Letters:
		return "letters"
	}
	return ""
}

type ExportPage struct {
	Filename    string
	IssueNumber int
	Title       string
	PageFrom    int
	PageTo      int
}
