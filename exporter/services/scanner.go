package services

import (
	contextPkg "context"
	"fmt"
	exporterApi "github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"golang.org/x/exp/maps"
	"sort"
)

type Scanner struct {
	storage *Storage
}

func toStories(issues []api.Issue) []*exporterApi.Story {
	storyMap := make(map[string]*exporterApi.Story)

	for _, issue := range issues {
		for _, episode := range issue.Episodes {
			// If the series - story combo exists, add to its episodes
			if story, ok := storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)]; ok {
				story.Episodes = append(story.Episodes, exporterApi.Episode{Episode: episode, Filename: issue.Filename, IssueNumber: issue.IssueNumber})
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
				s := exporterApi.Story{
					Title:  episode.Title,
					Series: episode.Series,
					Episodes: []exporterApi.Episode{{
						Episode:     episode,
						Filename:    issue.Filename,
						IssueNumber: issue.IssueNumber,
					}},
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

func (s *Scanner) Scan(paths []string, knownTitles, skipTitles []string) ([]*exporterApi.Story, error) {
	ctx, _ := context.WithLogger()

	// Create a scanner with the provided configuration
	scanner := scan.NewScanner(knownTitles, skipTitles)

	issues := make([]api.Issue, 0)
	for _, v := range paths {
		foundInPath, err := scanner.Dir(ctx, v, 0)
		if err != nil {
			return nil, fmt.Errorf("scanning directory %s: %w", v, err)
		}
		issues = append(issues, foundInPath...)
	}

	return toStories(issues), nil
}

// StartScan begins scanning directories and returns an operation that can be observed
func (s *Scanner) StartScan(paths []string, knownTitles, skipTitles []string) *ScanOperation {
	op := NewScanOperation()

	ctx, cancel := contextPkg.WithCancel(contextPkg.Background())
	op.cancel = cancel

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		// Check if cancelled
		select {
		case <-ctx.Done():
			_ = op.Error.Set("Scan cancelled")
			return
		default:
		}

		foundStories, err := s.Scan(paths, knownTitles, skipTitles)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Convert to untyped for binding
		untypedStories := make([]interface{}, len(foundStories))
		storiesToStore := make([]exporterApi.Story, len(foundStories))
		for i, v := range foundStories {
			untypedStories[i] = v
			storiesToStore[i] = *v
		}

		if err := op.Stories.Set(untypedStories); err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Store the stories
		if s.storage != nil {
			if err := s.storage.StoreStories(storiesToStore); err != nil {
				_ = op.Error.Set("Failed to save stories: " + err.Error())
			}
		}
	}()

	return op
}

func NewScanner(storage *Storage) *Scanner {
	return &Scanner{
		storage: storage,
	}
}
