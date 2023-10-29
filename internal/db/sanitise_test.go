package db

import (
	"fmt"
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

	results := []suggestionsResults{
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
		{
			Title: "Strontium Dug",
			Count: 1,
		},
		{
			Title: "Strontium Dog",
			Count: 15,
		},
	}

	t.Run("Sanitising test", func(t *testing.T) {
		suggestions := getSuggestions(appEnv, results, 0)

		if len(suggestions) != 2 {
			t.Errorf(fmt.Sprintf("sanitising test: expected %d suggestions, got %d", 3, len(suggestions)))
		}
		if !slices.Contains(suggestions, Suggestion{From: "Judge Fredd", To: "Judge Dredd"}) {
			t.Errorf(fmt.Sprintf("Suggestion to rename Judge Fredd not found"))
		}
	})

	// Add assertions to check if the function behaves as expected
	// This part is left as an exercise for the reader as it depends on the specific behavior of the function
}
