package scan

import (
	"context"
	"slices"
	"testing"

	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
)

func TestGetSuggestions(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		knownTitles    []string
		input          []*titleCounts
		suggestionType SuggestionType
		expectedOutput []Suggestion
	}{
		{
			name: "Sanitising test",
			input: []*titleCounts{
				{
					Title: "Judge Dredd",
					Count: 10,
				},
				{
					Title: "Judge Fredd",
					Count: 1,
				},
				{
					Title: "Brink",
					Count: 5,
				},
				{
					Title: "Renk",
					Count: 3,
				},
			},
			suggestionType: SeriesTitle,
			expectedOutput: []Suggestion{
				{From: "Judge Fredd", To: "Judge Dredd"},
			},
		},
		{
			name: "Dynamic levenshtein distance for short titles",
			input: []*titleCounts{
				{
					Title: "Brink",
					Count: 5,
				},
				{
					Title: "Renk",
					Count: 3,
				},
			},
			suggestionType: SeriesTitle,
			expectedOutput: []Suggestion{},
		},
		{
			name: "Strontium Dug",
			input: []*titleCounts{
				{
					Title: "Strontium Dug",
					Count: 1,
				},
				{
					Title: "Strontium Dog",
					Count: 15,
				},
			},
			suggestionType: SeriesTitle,
			expectedOutput: []Suggestion{
				{
					From: "Strontium Dug",
					To:   "Strontium Dog",
				},
			},
		},
		{
			name:        "Strontium Dug - preserved",
			knownTitles: []string{"Strontium Dug"},
			input: []*titleCounts{
				{
					Title: "Strontium Dug",
					Count: 1,
				},
				{
					Title: "Strontium Dog",
					Count: 15,
				},
			},
			suggestionType: SeriesTitle,
			expectedOutput: []Suggestion{},
		},
		{
			name:        "Brink - Hatebox",
			knownTitles: []string{},
			input: []*titleCounts{
				{
					Title: "Hate Box",
					Count: 6,
				},
				{
					Title: "Hatebox",
					Count: 2,
				},
			},
			suggestionType: EpisodeTitle,
			expectedOutput: []Suggestion{{
				From: "Hatebox",
				To:   "Hate Box",
				Type: EpisodeTitle,
			}},
		},
		{
			name:        "Sindex - Bulletopia",
			knownTitles: []string{},
			input: []*titleCounts{
				{
					Title:     "Bulletopia - Chapter One: Boys In The Hud",
					Count:     1,
					FirstSeen: 1,
					LastSeen:  1,
				},
				{
					Title:     "Boys In The Hud",
					Count:     1,
					FirstSeen: 2,
					LastSeen:  2,
				},
			},
			suggestionType: EpisodeTitle,
			expectedOutput: []Suggestion{{
				From: "Boys In The Hud",
				To:   "Bulletopia - Chapter One: Boys In The Hud",
				Type: EpisodeTitle,
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logger := logr.FromContextOrDiscard(context.TODO())
			suggestions := getSuggestions(logger, tc.knownTitles, tc.input, tc.suggestionType)

			if len(suggestions) != len(tc.expectedOutput) {
				t.Errorf("%s: expected %d suggestions, got %d", tc.name, len(tc.expectedOutput), len(suggestions))
			}

			for _, expectedSuggestion := range tc.expectedOutput {
				if !slices.Contains(suggestions, expectedSuggestion) {
					t.Errorf("%s: expected suggestion %v not found", tc.name, expectedSuggestion)
				}
			}
		})
	}
}

func TestDetectStorylineBookTitles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		issues []api.Issue
		checks []struct {
			issueIdx      int
			episodeIdx    int
			expectedTitle string
		}
	}{
		{
			name: "identical subtitles with different book numbers swap format",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Cold In The Bones", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: The Cold In The Bones", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "The Cold In The Bones: Book One"},
				{1, 0, "The Cold In The Bones: Book Two"},
			},
		},
		{
			name: "single book subtitle unchanged",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Cold In The Bones", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Book One: The Cold In The Bones"},
			},
		},
		{
			name: "different subtitles unchanged",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: Alpha", Part: 1},
						{Series: "Hershey", Title: "Book Two: Beta", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Book One: Alpha"},
				{0, 1, "Book Two: Beta"},
			},
		},
		{
			name: "non-book titles pass through",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Judge Dredd", Title: "Judge Dredd", Part: 1},
						{Series: "Strontium Dog", Title: "Strontium Dog", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Judge Dredd"},
				{0, 1, "Strontium Dog"},
			},
		},
		{
			name: "duplicate book number in different issues unchanged",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Story", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Story", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Book One: The Story"},
				{1, 0, "Book One: The Story"},
			},
		},
		{
			name: "three books all swap",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Trilogy", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: The Trilogy", Part: 1},
					},
				},
				{
					IssueNumber: 300,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Three: The Trilogy", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "The Trilogy: Book One"},
				{1, 0, "The Trilogy: Book Two"},
				{2, 0, "The Trilogy: Book Three"},
			},
		},
		{
			name: "two series, one triggers",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: Shared Story", Part: 1},
						{Series: "Other Series", Title: "Book One: Solo", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: Shared Story", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Shared Story: Book One"},
				{0, 1, "Book One: Solo"},
				{1, 0, "Shared Story: Book Two"},
			},
		},
		{
			name: "bare book title without subtitle unchanged",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Book One"},
			},
		},
		{
			name: "multiple episodes with same book word all swap",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Story", Part: 1},
					},
				},
				{
					IssueNumber: 150,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Story", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: The Story", Part: 1},
					},
				},
				{
					IssueNumber: 300,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: The Story", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "The Story: Book One"},
				{1, 0, "The Story: Book One"},
				{2, 0, "The Story: Book One"},
				{3, 0, "The Story: Book Two"},
			},
		},
		{
			name: "multiple episodes of both book words all swap",
			issues: []api.Issue{
				{
					IssueNumber: 100,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: Alpha", Part: 1},
					},
				},
				{
					IssueNumber: 200,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book One: Alpha", Part: 1},
					},
				},
				{
					IssueNumber: 300,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: Alpha", Part: 1},
					},
				},
				{
					IssueNumber: 400,
					Episodes: []*api.Episode{
						{Series: "Hershey", Title: "Book Two: Alpha", Part: 1},
					},
				},
			},
			checks: []struct {
				issueIdx      int
				episodeIdx    int
				expectedTitle string
			}{
				{0, 0, "Alpha: Book One"},
				{1, 0, "Alpha: Book One"},
				{2, 0, "Alpha: Book Two"},
				{3, 0, "Alpha: Book Two"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logger := logr.FromContextOrDiscard(context.TODO())
			detectStorylineBookTitles(&tc.issues, logger)
			for _, check := range tc.checks {
				got := tc.issues[check.issueIdx].Episodes[check.episodeIdx].Title
				if got != check.expectedTitle {
					t.Errorf("%s: issue[%d].episode[%d] expected %q, got %q",
						tc.name, check.issueIdx, check.episodeIdx, check.expectedTitle, got)
				}
			}
		})
	}
}
