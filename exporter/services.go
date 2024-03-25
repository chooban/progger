package exporter

import (
	"cmp"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"golang.org/x/exp/maps"
	"path/filepath"
	"slices"
	"sort"
)

type Exporter struct {
	BoundSourceDir binding.String
	BoundExportDir binding.String
}

func (e *Exporter) Export(stories []*Story) error {
	sourceDir, err := e.BoundSourceDir.Get()
	if err != nil {
		return err
	}
	exportDir, err := e.BoundExportDir.Get()
	if err != nil {
		return err
	}

	toExport := make([]api.ExportPage, 0)
	for _, story := range stories {
		//story := v.(*Story)
		if story.ToExport {
			for _, e := range story.Episodes {
				toExport = append(toExport, api.ExportPage{
					Filename:    filepath.Join(sourceDir, e.Filename),
					PageFrom:    e.FirstPage,
					PageTo:      e.LastPage,
					IssueNumber: e.IssueNumber,
					Title:       fmt.Sprintf("%s - Part %d", e.Title, e.Part),
				})
			}
		}
	}
	if len(toExport) == 0 {
		println("Nothing to export")
		return nil
	}
	// Sort by issue number. We sometimes have issues being wrongly grouped, but surely we never want anything
	// other than issue order?
	slices.SortFunc(toExport, func(i, j api.ExportPage) int {
		return cmp.Compare(i.IssueNumber, j.IssueNumber)
	})

	// Do the export
	err = scan.Build(WithLogger(), toExport, filepath.Join(exportDir, "export.pdf"))
	if err != nil {
		return err
	}

	return nil
}

type Scanner struct {
	IsScanning   binding.Bool
	BoundStories binding.UntypedList
}

func toStories(issues []api.Issue) []*Story {
	storyMap := make(map[string]*Story)

	for _, issue := range issues {
		for _, episode := range issue.Episodes {
			// If the series - story combo exists, add to its episodes
			if story, ok := storyMap[fmt.Sprintf("%s - %s", episode.Series, episode.Title)]; ok {
				story.Episodes = append(story.Episodes, Episode{episode, issue.Filename, issue.IssueNumber})
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
				s := Story{
					Title:      episode.Title,
					Series:     episode.Series,
					Episodes:   []Episode{{episode, issue.Filename, issue.IssueNumber}},
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

func (s *Scanner) Scan(path string) {
	s.IsScanning.Set(true)
	ctx := WithLogger()
	issues := scan.Dir(ctx, path, 0)

	stories := toStories(issues)
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
		BoundStories: binding.NewUntypedList(),
	}
}

func NewExporter(src, export binding.String) *Exporter {
	return &Exporter{
		src, export,
	}
}
