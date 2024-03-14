package scan

import (
	"context"
	"github.com/go-logr/logr"
	"slices"
	"testing"
)

func TestGetSuggestions(t *testing.T) {

	testCases := []struct {
		name           string
		knownTitles    []string
		input          []*titleCounts
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
			expectedOutput: []Suggestion{{
				From: "Hatebox",
				To:   "Hate Box",
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logr.FromContextOrDiscard(context.TODO())
			suggestions := getSuggestions(logger, tc.knownTitles, tc.input, 0)

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
