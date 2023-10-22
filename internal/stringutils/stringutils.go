package stringutils

import (
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"regexp"
	"strconv"
	"strings"
)

func ParseTextNumber(textNum string) (part int, err error) {
	part, convErr := strconv.Atoi(textNum)

	if convErr == nil {
		return part, nil
	}

	switch strings.ToLower(strings.TrimSpace(textNum)) {
	case "one":
		part = 1
	case "two":
		part = 2
	case "three":
		part = 3
	case "four":
		part = 4
	case "five":
		part = 5
	case "six":
		part = 6
	case "seven":
		part = 7
	case "eight":
		part = 8
	case "nine":
		part = 9
	case "ten":
		part = 10
	default:
		part = 0
		err = errors.New("default value returned")
	}
	return
}

func TrimNonAlphaNumeric(input string) string {
	patternTrailing := "[^a-zA-Z0-9]+$"
	patternLeading := "^[^a-zA-Z0-9]+"
	re := regexp.MustCompile(patternTrailing)
	reLeading := regexp.MustCompile(patternLeading)
	return re.ReplaceAllString(reLeading.ReplaceAllString(input, ""), "")
}

func CapitalizeWords(sentence string) string {
	skipWords := map[string]string{
		"Of":       "of",
		"Vs":       "vs",
		"3Rillers": "3rillers",
	}
	capitalized := cases.Title(language.BritishEnglish).String(sentence)

	for k, v := range skipWords {
		capitalized = strings.ReplaceAll(capitalized, k, v)
	}

	return capitalized
}

func ContainsI(s string, substr string) bool {
	re := regexp.MustCompile(fmt.Sprintf("(?i)\\b%s\\b", substr))
	return re.MatchString(s)
}

func FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	subexpNames := regex.SubexpNames()
	for i, name := range match {
		if len(subexpNames[i]) == 0 {
			continue
		}
		results[subexpNames[i]] = name
	}
	return results
}
