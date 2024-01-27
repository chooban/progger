package stringutils

import (
	"reflect"
	"regexp"
	"testing"
)

func TestParseTextNumber(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
	}{
		{"One", "one", 1},
		{"Two", "two", 2},
		{"Three", "three", 3},
		{"Four", "four", 4},
		{"Five", "five", 5},
		{"Six", "six", 6},
		{"Seven", "seven", 7},
		{"Eight", "eight", 8},
		{"Nine", "nine", 9},
		{"Ten", "ten", 10},
		{"Eleven", "eleven", 11},
		{"Twelve", "twelve", 12},
		{"Thirteen", "thirteen", 13},
		{"Fourteen", "fourteen", 14},
		{"Fifteen", "fifteen", 15},
		{"Sixteen", "sixteen", 16},
		{"Seventeen", "seventeen", 17},
		{"Eighteen", "eighteen", 18},
		{"Nineteen", "nineteen", 19},
		{"Twenty", "twenty", 20},
		{"Invalid", "invalid", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := ParseTextNumber(tc.input)
			if got != tc.expected {
				t.Errorf("ParseTextNumber(%v) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestTrimNonAlphaNumeric(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Start", "#Hello", "Hello"},
		{"End", "Hello#", "Hello"},
		{"Both", "#Hello#", "Hello"},
		{"None", "Hello", "Hello"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := TrimNonAlphaNumeric(tc.input)
			if got != tc.expected {
				t.Errorf("TrimNonAlphaNumeric(%v) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestCapitalizeWords(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Single", "hello", "Hello"},
		{"Multiple", "hello world", "Hello World"},
		{"None", "", ""},
		{"Skip words", "world of goo", "World of Goo"},
		{name: "MC-1", input: "mega-city one", expected: "Mega-City One"},
		{name: "Robo-Hunter", input: "judge dredd vs robo-hunter", expected: "Judge Dredd vs Robo-Hunter"},
		{name: "letter after number", input: "3rillers", expected: "3rillers"},
		{name: "ABC", input: "Abc", expected: "ABC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := CapitalizeWords(tc.input)
			if got != tc.expected {
				t.Errorf("CapitalizeWords(%v) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestContainsI(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"Present", "Hello", "hello", true},
		{"NotPresent", "Hello", "world", false},
		{"Case", "Hello", "HELLO", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsI(tc.s, tc.substr)
			if got != tc.expected {
				t.Errorf("ContainsI(%v, %v) = %v; want %v", tc.s, tc.substr, got, tc.expected)
			}
		})
	}
}

func TestFindNamedMatches(t *testing.T) {
	testCases := []struct {
		name     string
		regex    *regexp.Regexp
		str      string
		expected map[string]string
	}{
		{
			"Match",
			regexp.MustCompile(`(?P<First>\w+)\s(?P<Last>\w+)`),
			"Hello World",
			map[string]string{"First": "Hello", "Last": "World"},
		},
		{
			"NoMatch",
			regexp.MustCompile(`(?P<First>\w+)\s(?P<Last>\w+)`),
			"HelloWorld",
			map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindNamedMatches(tc.regex, tc.str)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("FindNamedMatches(%v) = %v; want %v", tc.str, got, tc.expected)
			}
		})
	}
}
