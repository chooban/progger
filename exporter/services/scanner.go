package services

import (
	"context"
	"fmt"
	api2 "github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"golang.org/x/exp/maps"
	"sort"
)

type Scanner struct {
	ctxt context.Context
}

func toStories(issues []api.Issue) []*api2.Story {
	storyMap := make(map[string]*api2.Story)

	for _, issue := range issues {
		for _, episode := range issue.Episodes {
			// If the series - story combo exists, add to its episodes
			if story, ok := storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)]; ok {
				story.Episodes = append(story.Episodes, api2.Episode{Episode: episode, Filename: issue.Filename, IssueNumber: issue.IssueNumber})
				sort.Slice(story.Episodes, func(i, j int) bool {
					return story.Episodes[i].IssueNumber < story.Episodes[j].IssueNumber
				})
				story.Issues = append(story.Issues, issue.IssueNumber)
				if issue.IssueNumber < story.FirstIssue {
					story.FirstIssue = issue.IssueNumber
				}
				if issue.IssueNumber > story.LastIssue {
					story.LastIssue = issue.IssueNumber
				}
			} else {
				s := api2.Story{
					Title:      episode.Title,
					Series:     episode.Series,
					Episodes:   []api2.Episode{{episode, issue.Filename, issue.IssueNumber}},
					FirstIssue: issue.IssueNumber,
					LastIssue:  issue.IssueNumber,
					Issues:     []int{issue.IssueNumber},
					ToExport:   false,
				}
				storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)] = &s
			}
		}
	}

	stories := maps.Values(storyMap)
	sort.Slice(stories, func(i, j int) bool {
		storyI := stories[i]
		storyJ := stories[j]

		if storyI.Series != storyJ.Series {
			return storyI.Series < storyJ.Series
		}
		return stories[i].FirstIssue < stories[j].FirstIssue
	})

	return stories
}

func (s *Scanner) Scan(path string) []*api2.Story {
	issues := scan.Dir(s.ctxt, path, 0)

	return toStories(issues)
}

func NewScanner(ctx context.Context) *Scanner {
	return &Scanner{
		ctxt: ctx,
	}
}
