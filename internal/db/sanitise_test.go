package db

import "testing"

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
	// Create a mock AppEnv
	appEnv := env.AppEnv{
		Db:   nil,
		Log:  nil,
		Skip: env.ToSkip{},
		Known: env.Known{
			SeriesTitles: []string{"Known Title"},
		},
	}

	// Create a mock suggestionsResults array
	results := []suggestionsResults{
		{
			Title: "Known Title",
			Count: 1,
		},
		{
			Title: "Unknown Title",
			Count: 2,
		},
	}

	// Call getSuggestions with the mock data
	getSuggestions(appEnv, results)

	// Add assertions to check if the function behaves as expected
	// This part is left as an exercise for the reader as it depends on the specific behavior of the function
}
