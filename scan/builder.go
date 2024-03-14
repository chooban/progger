package scan

import (
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal/pdf"
	"github.com/chooban/progger/scan/internal/stringutils"
	"github.com/divan/num2words"
	"github.com/go-logr/logr"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func getProgNumber(inFile string) (int, error) {
	filename := filepath.Base(inFile)
	regex := regexp.MustCompile(`(\b[^()])(?P<issue>\d{1,4})(\b[^()])`)

	namedResults := stringutils.FindNamedMatches(regex, filename)
	if len(namedResults) > 0 {
		return strconv.Atoi(stringutils.TrimNonAlphaNumeric(namedResults["issue"]))
	}
	return 0, errors.New("no number found in filename")
}

func buildIssue(log logr.Logger, filename string, details []pdf.EpisodeDetails, knownTitles []string, skipTitles []string) api.Issue {
	issueNumber, _ := getProgNumber(filename)
	allEpisodes := make([]*api.Episode, 0)

	for _, d := range details {
		b := d.Bookmark
		log.V(2).Info(fmt.Sprintf("Extracting details from %s", b.Title))
		part, series, title := extractDetailsFromPdfBookmark(b.Title)

		if series == "" {
			log.V(1).Info(fmt.Sprintf("Odd title: %s", b.Title))
			continue
		}
		// Check to see if the series is close to any of the blessed titles
		for _, v := range knownTitles {
			if series == v {
				break
			}
			distance := levenshtein.DistanceForStrings(
				[]rune(strings.ToLower(v)),
				[]rune(strings.ToLower(series)),
				levenshtein.DefaultOptions,
			)
			log.V(2).Info(fmt.Sprintf("Distance between '%s' and '%s' is %d", v, series, distance))
			if distance < 5 {
				series = v
			}
		}

		if shouldIncludeEpisode(log, skipTitles, series, title) {
			log.V(1).Info(fmt.Sprintf("Extracting creators from %s", d.Credits))
			credits := extractCreatorsFromCredits(d.Credits)

			allEpisodes = append(allEpisodes, &api.Episode{
				Title:     title,
				Series:    series,
				Part:      part,
				FirstPage: b.PageFrom,
				LastPage:  b.PageThru,
				Credits:   credits,
			})
		} else {
			log.V(1).Info(fmt.Sprintf("Skipping. Series: %s. Episode: %s", series, title))
		}
	}
	issue := api.Issue{
		Publication: "2000 AD",
		IssueNumber: issueNumber,
		Filename:    filepath.Base(filename),
		Episodes:    allEpisodes,
	}

	return issue
}

func extractDetailsFromPdfBookmark(bookmarkTitle string) (episodeNumber int, series string, storyline string) {
	// We don't want any zero parts. It's 1 if not specified
	episodeNumber = -1

	// Very rarely, someone decides to use a number for a book when most are words
	bookRegex := regexp.MustCompile(`(?i)book \d+`)
	bookmarkTitle = bookRegex.ReplaceAllStringFunc(bookmarkTitle, func(s string) string {
		parts := strings.Split(s, " ")
		num, _ := strconv.Atoi(parts[1])

		// Put it back, but with an extra colon in there. Some of the `Book X` bookmarks don't have one, and this
		// messes things up. If we put one in we might split twice, but then we remove the empty strings from the
		// array.
		return fmt.Sprintf(":%s %s", parts[0], num2words.Convert(num))
	})

	splitRegex := regexp.MustCompile(`([:_"()]|(- )|\.{3})`)
	parts := splitRegex.Split(bookmarkTitle, -1)
	parts = slices.DeleteFunc(parts, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	if len(parts) == 3 {
		// Three-way split? Series, storyline, episodeNumber
		series = stringutils.CapitalizeWords(strings.TrimSpace(parts[0]))
		storyline = stringutils.CapitalizeWords(strings.TrimSpace(parts[1]))
		episodeNumber = extractPartNumberFromString(parts[2])

		return
	}

	partFinder := regexp.MustCompile(`(?i)^.*(?P<whole>part (?P<episodeNumber>\w+)).*$`)
	if partFinder.MatchString(bookmarkTitle) {
		namedResults := stringutils.FindNamedMatches(partFinder, bookmarkTitle)
		partString := namedResults["episodeNumber"]
		maybePart, err := stringutils.ParseTextNumber(partString)
		if err == nil {
			episodeNumber = maybePart
		}
		toReplace := regexp.MustCompile("(?i)\\s+part " + partString + "[^a-zA-Z0-9]*")
		bookmarkTitle = toReplace.ReplaceAllString(bookmarkTitle, " ")
	}

	titleSplit := splitRegex.Split(bookmarkTitle, -1)
	titleSplit = slices.DeleteFunc(titleSplit, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	series = strings.TrimSpace(titleSplit[0])
	if len(titleSplit) > 2 {
		// Already set, so we must have had "Part Two" somewhere else. Just put it all back together and call
		// it a storyline
		trimmedParts := make([]string, len(titleSplit)-1)
		for _, v := range titleSplit[1:] {
			trimmedParts = append(trimmedParts, strings.TrimSpace(v))
		}
		storyline = strings.Join(trimmedParts, ": ")
	} else {
		series = titleSplit[0]
		if len(titleSplit) > 1 {
			storyline = titleSplit[1]
		} else {
			// Assume eponymous
			storyline = series
		}
	}

	// At the end we set the default
	if episodeNumber == -1 {
		episodeNumber = 1
	}
	series = stringutils.TrimNonAlphaNumeric(stringutils.CapitalizeWords(series))
	storyline = stringutils.TrimNonAlphaNumeric(stringutils.CapitalizeWords(storyline))

	return
}

func shouldIncludeEpisode(logger logr.Logger, seriesToSkip []string, seriesTitle string, episodeTitle string) bool {
	pagesToSkip := []string{
		"Star scan",
		"Normal Opti",
		"Pin up",
		"Pin-up",
		"Cover",
		"Nerve Centre",
		"Input",
		"Art Stars",
		"Art Print",
		"Tharg interlude",
		"Thrill-search",
		"Thought Bubble",
		"Insight profile",
		"How to draw",
		"Feature",
		"Brimful of thrills",
		"In Memoriam",
	}

	for _, s := range seriesToSkip {
		if seriesTitle == s {
			logger.Info(fmt.Sprintf("Skipping series %s", s))
			return false
		}
	}
	for _, s := range pagesToSkip {
		for _, t := range []string{episodeTitle, seriesTitle} {
			if stringutils.ContainsI(t, s) || levenshtein.DistanceForStrings([]rune(s), []rune(t), levenshtein.DefaultOptions) < 5 {
				logger.V(1).Info(fmt.Sprintf("%s contains, or is close to, %s", t, s))
				return false
			}
		}
	}
	return true
}

func extractPartNumberFromString(toParse string) (part int) {
	part = 1
	toParse = strings.ToLower(toParse)
	if strings.Contains(toParse, "part") {
		toParse = strings.TrimSpace(strings.Split(toParse, "part")[1])
	}
	maybePart, err := strconv.Atoi(strings.TrimSpace(toParse))
	if err != nil {
		if maybePart, err = stringutils.ParseTextNumber(toParse); err == nil {
			part = maybePart
		}
	} else {
		part = maybePart
	}
	return
}

func extractCreatorsFromCredits(toParse string) (credits api.Credits) {
	credits = api.Credits{}

	var currentRole = api.Unknown
	var tokens = strings.Split(toParse, " ")
	currentCreatorString := make([]string, 0)
	for _, t := range tokens {
		if t == "" {
			continue
		}
		r, err := api.NewRole(strings.ToLower(t))
		if currentRole != api.Unknown && err != nil {
			currentCreatorString = append(currentCreatorString, strings.TrimSpace(t))
		} else if r != currentRole && err == nil {
			if currentRole == api.Unknown {
				currentRole = r
				continue
			}
			credits[currentRole] = normaliseCreators(currentCreatorString)

			// Zero the string
			currentCreatorString = currentCreatorString[:0]
			currentRole = r
		}
	}
	credits[currentRole] = normaliseCreators(currentCreatorString)

	return credits
}

func normaliseCreators(input []string) []string {
	tokens := strings.Split(stringutils.CapitalizeWords(strings.Join(input, " ")), "&")

	creators := make([]string, 0)
	for _, v := range tokens {
		creators = append(creators, strings.TrimSpace(v))
	}
	return creators
}
