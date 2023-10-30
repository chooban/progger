package db

import (
	"github.com/chooban/progdl-go/internal/env"
	"slices"
	"testing"
)

func TestGetTargetLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedTarget int
	}{
		{
			name:           "Renk",
			input:          "Renk",
			expectedTarget: 1,
		},
		{
			name:           "Hook-Jaw",
			input:          "Hook Jaw",
			expectedTarget: 2,
		},
		{
			name:           "Anderson",
			input:          "Anderson, Psi-Division",
			expectedTarget: 5,
		},
		{
			name:           "Brink",
			input:          "Robohunter",
			expectedTarget: 3,
		},
		{
			name:           "Judge Dredd",
			input:          "Judge Dredd",
			expectedTarget: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotDistance := getTargetLevenshteinDistance(tc.input)
			if gotDistance != tc.expectedTarget {
				t.Errorf("getTargetLevenshteinDistance(%v) = %v; want %v", tc.input, gotDistance, tc.expectedTarget)
			}
		})
	}
}

func TestGetSuggestions(t *testing.T) {
	appEnv := createAppEnv()

	testCases := []struct {
		name           string
		appEnv         env.AppEnv
		input          []suggestionsResults
		expectedOutput []Suggestion
	}{
		{
			name:   "Sanitising test",
			appEnv: appEnv,
			input: []suggestionsResults{
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
			name:   "Dynamic levenshtein distance for short titles",
			appEnv: appEnv,
			input: []suggestionsResults{
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
			name:   "Strontium Dug",
			appEnv: appEnv,
			input: []suggestionsResults{
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
			name: "Strontium Dug - preserved",
			appEnv: func() env.AppEnv {
				appEnv := createAppEnv()
				appEnv.Known.SeriesTitles = []string{"Strontium Dug"}
				return appEnv
			}(),
			input: []suggestionsResults{
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
		// Add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suggestions := getSuggestions(tc.appEnv, tc.input, 0)

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
