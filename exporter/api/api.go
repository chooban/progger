package api

import (
	"fmt"
	"github.com/chooban/progger/download"
	scanApi "github.com/chooban/progger/scan/api"
	"slices"
	"strconv"
	"strings"
)

type Episode struct {
	*scanApi.Episode
	Issue       *download.DigitalComic
	Filename    string
	IssueNumber int
}

type Story struct {
	Title      string
	Series     string
	Episodes   []Episode
	FirstIssue int
	LastIssue  int
	Issues     []int
	ToExport   bool
}

func (s *Story) Display() string {
	return strings.Join([]string{s.Series, s.Title}, " - ")
}

func (s *Story) IssueSummary() string {
	if len(s.Issues) == 1 {
		return strconv.Itoa(s.Issues[0])
	}
	toSort := slices.Clone(s.Issues)
	slices.Sort(toSort)

	progs := []string{}
	start := 0
	for i, v := range toSort {
		if i == 0 {
			continue
		}
		if toSort[i-1] == v-1 {
			continue
		}
		progs = append(progs, fmt.Sprintf("%d - %d", toSort[start], toSort[i-1]))
		start = i
	}

	if toSort[start] == toSort[len(toSort)-1] {
		progs = append(progs, strconv.Itoa(toSort[start]))
	} else {
		progs = append(progs, fmt.Sprintf("%d - %d", toSort[start], toSort[len(toSort)-1]))
	}

	return strings.Join(progs, ", ")
}

type Downloadable struct {
	Comic      download.DigitalComic
	Downloaded bool
}
