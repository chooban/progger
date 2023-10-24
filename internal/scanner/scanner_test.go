package scanner

import (
	"github.com/chooban/progdl-go/internal/db"
	"testing"
)

func TestGetProgNumber(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedNumber int
	}{
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
			gotNumber := getProgNumber(tc.input, nil)
			if gotNumber != tc.expectedNumber {
				t.Errorf("getProgNumber(%v) = %v; want %v", tc.input, gotNumber, tc.expectedNumber)
			}
		})
	}
}
func TestShouldIncludeIssue(t *testing.T) {
	testCases := []struct {
		name     string
		input    db.Issue
		expected bool
	}{
		{
			name:     "Issue with number 0",
			input:    db.Issue{IssueNumber: 0},
			expected: false,
		},
		{
			name:     "Issue with number 1234",
			input:    db.Issue{IssueNumber: 1234},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldIncludeIssue(tc.input)
			if got != tc.expected {
				t.Errorf("shouldIncludeIssue(%v) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}
