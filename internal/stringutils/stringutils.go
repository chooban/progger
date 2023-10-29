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
	case "eleven":
		part = 11
	case "twelve":
		part = 12
	case "thirteen":
		part = 13
	case "fourteen":
		part = 14
	case "fifteen":
		part = 15
	case "sixteen":
		part = 16
	case "seventeen":
		part = 17
	case "eighteen":
		part = 18
	case "nineteen":
		part = 19
	case "twenty":
		part = 20
	default:
		part = 0
		err = errors.New("default value returned")
	}
	return
}

func TrimNonAlphaNumeric(input string) string {
	patternTrailing := "[^a-zA-Z0-9!\\.]+$"
	patternLeading := "^[^a-zA-Z0-9']+"
	re := regexp.MustCompile(patternTrailing)
	reLeading := regexp.MustCompile(patternLeading)
	return re.ReplaceAllString(reLeading.ReplaceAllString(input, ""), "")
}

func CapitalizeWords(sentence string) string {
	skipWords := map[string]string{
		"Of":        "of",
		"Vs":        "vs",
		"3Rillers":  "3rillers",
		"s.t.a.r.s": "S.t.a.r.s",
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
