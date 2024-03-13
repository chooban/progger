package exporter

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"golang.org/x/exp/maps"
	"sort"
)

type Scanner struct {
	IsScanning   binding.Bool
	Issues       []api.Issue
	BoundIssues  binding.UntypedList
	BoundStories binding.UntypedList
}

type Story struct {
	Title      string
	Series     string
	Episodes   []api.Episode
	FirstIssue int
	LastIssue  int
}

func toStories(issues []api.Issue) []*Story {
	stories := make([]*Story, len(issues))

	storyMap := make(map[string]*Story)

	for _, issue := range issues {
		for _, episode := range issue.Episodes {
			// If the series - story combo exists, add to its episodes
			if story, ok := storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)]; ok {
				story.Episodes = append(story.Episodes, episode)
				sort.Slice(story.Episodes, func(i, j int) bool {
					return story.Episodes[i].Part < story.Episodes[j].Part
				})
				if issue.IssueNumber < story.FirstIssue {
					story.FirstIssue = issue.IssueNumber
				}
				if issue.IssueNumber > story.LastIssue {
					story.LastIssue = issue.IssueNumber
				}
			} else {
				s := Story{
					Title:      episode.Title,
					Series:     episode.Series,
					Episodes:   []api.Episode{episode},
					FirstIssue: issue.IssueNumber,
					LastIssue:  issue.IssueNumber,
				}
				storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)] = &s
			}
		}
	}
	stories = maps.Values(storyMap)
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

func (s *Scanner) Scan(path string) {
	s.IsScanning.Set(true)
	ctx := WithLogger()
	issues := scan.Dir(ctx, path, 0)
	s.Issues = issues

	for _, v := range issues {
		s.BoundIssues.Append(v)
	}

	stories := toStories(s.Issues)
	for _, v := range stories {
		s.BoundStories.Append(v)
	}
	if err := s.IsScanning.Set(false); err != nil {
		println(err.Error())
	}
}

func NewScanner() *Scanner {
	isScanning := binding.NewBool()

	return &Scanner{
		IsScanning:   isScanning,
		Issues:       []api.Issue{},
		BoundIssues:  binding.NewUntypedList(),
		BoundStories: binding.NewUntypedList(),
	}
}
