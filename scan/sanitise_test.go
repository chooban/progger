package scan

import (
	"context"
	"slices"
	"testing"

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
