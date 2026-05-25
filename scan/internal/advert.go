package internal

import (
	"regexp"
	"strings"
)

var advertPattern = regexp.MustCompile(`on sale \d{1,2} \w+ \d{4}`)
var isbnPattern = regexp.MustCompile(`\d{3}-\d-\d{5}-\d{3}-\d`)
var advertTexts = []string{
	"on sale now",
	"shop.",
}

func IsAdvertPageText(text string) bool {
	t := strings.ToLower(text)

	return strings.Contains(t, "on sale now") || advertPattern.MatchString(t) || isbnPattern.MatchString(t)
}

func TrimAdvertPages(pageFrom, pageTo int, isAdvert func(pageIndex int) bool) int {
	if pageFrom > pageTo {
		return pageTo
	}

	allAdverts := true
	allClean := true
	firstAdvert := 0

	for pageIndex := pageFrom; pageIndex <= pageTo; pageIndex++ {
		if isAdvert(pageIndex) {
			allClean = false
			if firstAdvert == 0 {
				firstAdvert = pageIndex
			}
		} else {
			allAdverts = false
		}
	}

	if allAdverts || allClean {
		return pageTo
	}

	return firstAdvert - 1
}
