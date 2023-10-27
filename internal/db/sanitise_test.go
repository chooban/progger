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
