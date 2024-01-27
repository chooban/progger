package internal

import (
	"testing"
)

func TestGetProgNumber(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedNumber int
	}{
		{
			name:           "Prog 123",
			input:          "2000AD 123 (1977).pdf",
			expectedNumber: 123,
		},
		{
			name:           "Prog 2000",
			input:          "2000AD 2000 (1977).pdf",
			expectedNumber: 2000,
		},
		{
			name:           "Prog 1234",
			input:          "2000AD 1234 (1977).pdf",
			expectedNumber: 1234,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotNumber, _ := getProgNumber(tc.input)
			if gotNumber != tc.expectedNumber {
				t.Errorf("getProgNumber(%v) = %v; want %v", tc.input, gotNumber, tc.expectedNumber)
			}
		})
	}
}
