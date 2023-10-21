package stringutils

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
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
	skipWords := []string{"of"}
	words := strings.Fields(sentence) // Split the sentence into words
	var capitalizedWords []string

	for _, word := range words {
		if len(word) == 0 {
			continue
		}
		if slices.Contains(skipWords, word) {
			capitalizedWords = append(capitalizedWords, word)
		} else {
			// Capitalize the first letter of each word
			capitalizedWord := strings.ToUpper(word[:1]) + word[1:]
			capitalizedWords = append(capitalizedWords, capitalizedWord)
		}
	}

	// Join the capitalized words to form the final sentence
	capitalizedSentence := strings.Join(capitalizedWords, " ")
	return capitalizedSentence
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
