package internal

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chooban/progger/scan/api"
	"github.com/divan/num2words"
	"github.com/go-logr/logr"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func extractWords(text string) []string {
	smallWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true,
		"of": true, "in": true, "to": true, "for": true,
		"with": true, "on": true, "at": true, "by": true,
		"or": true, "as": true, "one": true,
	}

	re := regexp.MustCompile(`'[a-zA-Z]+$|'[a-zA-Z]+\b`)

	words := strings.Fields(text)
	result := make([]string, 0)
	for _, word := range words {
		cleanedWord := strings.ToLower(word)
		cleanedWord = re.ReplaceAllString(cleanedWord, "")
		cleanedWord = strings.TrimFunc(cleanedWord, func(r rune) bool {
			return !(unicode.IsLetter(r) || unicode.IsDigit(r))
		})
		cleanedWord, _ = strings.CutSuffix(cleanedWord, "'s")
		cleanedWord, _ = strings.CutSuffix(cleanedWord, "’s")
		if cleanedWord != "" && !smallWords[cleanedWord] && len(cleanedWord) > 2 {
			result = append(result, cleanedWord)
		}
	}
	return result
}

// findCoverDate attempts to find a date string in the cover text and then parse and convert it into
// the YYYY-MM-DD format
func findCoverDate(log logr.Logger, coverText string) string {
	var r = regexp.MustCompile(`(\d{2} \w{3,4} \d{2,4})`)
	var m = r.FindStringSubmatch(coverText)

	if len(m) > 0 && m[0] != "" {
		parts := strings.Split(m[0], " ")
		day, month, year := "02", "Jan", "06"
		if len(parts[1]) > 3 {
			month = "January"
		}
		if len(parts[2]) > 2 {
			year = "2006"
		}
		if t, err := time.Parse(fmt.Sprintf("%s %s %s", day, month, year), m[0]); err == nil {
			coverDate := t.Format("2006-01-02")
			return coverDate
		}
	}
	println(fmt.Sprintf("Cover text: %s", coverText))
	return ""
}

func findBestMatchingSeries(log logr.Logger, coverText string, episodes []*api.Episode) string {
	if len(episodes) == 0 {
		log.V(1).Info("No episodes to compare cover text against")
		return ""
	}

	log.V(0).Info("Breaking cover text into words", "coverText", coverText)
	coverWords := extractWords(coverText)
	if len(coverWords) == 0 {
		log.V(1).Info("No cover words to compare against")
		return ""
	}

	log.V(0).Info("Found cover words to compare against", "words", coverWords)

	processedSeries := make(map[string]bool)
	seriesScores := make(map[string]int)

	for _, ep := range episodes {
		seriesName := ep.Series
		if seriesName == "" {
			continue
		}
		if processedSeries[seriesName] {
			continue
		}
		processedSeries[seriesName] = true

		seriesWords := extractWords(seriesName)
		overlap := 0
		seriesWordSet := make(map[string]bool)
		for _, word := range seriesWords {
			seriesWordSet[word] = true
		}

		for _, coverWord := range coverWords {
			if seriesWordSet[coverWord] {
				overlap++
			}
		}

		seriesScores[seriesName] = overlap
	}

	if len(seriesScores) == 0 {
		log.V(1).Info("No series scored")
		return ""
	}

	var maxScore int
	var winningSeries []string

	for series, score := range seriesScores {
		if score > maxScore {
			maxScore = score
			winningSeries = []string{series}
		} else if score == maxScore && score > 0 {
			winningSeries = append(winningSeries, series)
		}
	}

	if len(winningSeries) == 1 {
		return winningSeries[0]
	}

	log.V(1).Info("No winning series found")

	// In this case, check for the edge case of "new thrill" and find a series that's part 1
	if slices.Contains(coverWords, "new") && slices.Contains(coverWords, "thrill") {
		for _, ep := range episodes {
			if ep.Part == 1 {
				return ep.Series
			}
		}
	}

	if slices.Contains(coverWords, "dexter") || slices.Contains(coverWords, "downlode") || slices.Contains(coverWords, "suzi") {
		for _, ep := range episodes {
			if ep.Series == "Azimuth" {
				return ep.Series
			}
		}

	}
	return ""
}

func getProgNumber(inFile string) (int, error) {
	filename := filepath.Base(inFile)

	knownFileNames := []*regexp.Regexp{
		regexp.MustCompile(`(\b[^()])(?P<issue>\d{1,4})(\b[^()])`),
		regexp.MustCompile(`PRG(?P<issue>\d{1,4})D`),
	}

	for _, regex := range knownFileNames {
		namedResults := FindNamedMatches(regex, filename)
		if len(namedResults) > 0 {
			return strconv.Atoi(TrimNonAlphaNumeric(namedResults["issue"]))
		}
	}
	return 0, errors.New("no number found in filename")
}

func BuildIssue(log logr.Logger, filename string, details []EpisodeDetails, coverText, indexText string, knownTitles []string, skipTitles []string) api.Issue {
	issueNumber, err := getProgNumber(filename)
	if err != nil {
		log.Error(err, "Error getting issue number")
		return api.Issue{}
	}
	allEpisodes := make([]*api.Episode, 0)

	// Get PDF page count for fixing invalid page ranges
	reader := NewPdfiumReader(log)
	pdfPageCount := 0
	if pageCount, err := reader.PageCount(filename); err == nil {
		pdfPageCount = pageCount
	}

	for _, d := range details {
		bookmark := d.Bookmark
		log.V(2).Info(fmt.Sprintf("Extracting details from %s", bookmark.Title))
		part, series, title := extractDetailsFromPdfBookmark(bookmark.Title)

		if series == "" {
			log.V(1).Info(fmt.Sprintf("Odd title: %s", bookmark.Title))
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
			credits := ExtractCreatorsFromCredits(d.Credits)

			// Fix page ranges: if LastPage is 0, it means "to end of PDF"
			pageFrom := bookmark.PageFrom
			pageTo := bookmark.PageThru
			if pageTo == 0 && pdfPageCount > 0 {
				pageTo = pdfPageCount
				log.Info("Fixed page range from PDF bookmark", "series", series, "title", title, "part", part, "pageFrom", pageFrom, "pageTo", pageTo)
			}

			allEpisodes = append(allEpisodes, &api.Episode{
				Title:     title,
				Series:    series,
				Part:      part,
				FirstPage: pageFrom,
				LastPage:  pageTo,
				Credits:   credits,
			})
		} else {
			log.V(1).Info(fmt.Sprintf("Skipping. Series: %s. Episode: %s", series, title))
		}
	}

	coverArtist := ""
	var splits = strings.Split(indexText, "\n")
	for i := 0; i < len(splits); i++ {
		if strings.Contains(strings.ToLower(splits[i]), "cover art") {
			coverArtist = strings.TrimSpace(splits[i+1])
		}
	}

	bestSeries := findBestMatchingSeries(log, coverText, allEpisodes)
	coverDate := findCoverDate(log, coverText)

	cover := api.Cover{
		Series:   bestSeries,
		Text:     coverText,
		Artist:   coverArtist,
		Filename: filename,
	}
	issue := api.Issue{
		Publication: "2000 AD",
		IssueNumber: issueNumber,
		Filename:    filename,
		Episodes:    allEpisodes,
		Cover:       cover,
		CoverDate:   coverDate,
	}

	return issue
}

func extractDetailsFromPdfBookmark(bookmarkTitle string) (episodeNumber int, series string, storyline string) {
	// We don't want any zero parts. It's 1 if not specified
	episodeNumber = -1

	// Very rarely, someone decides to use a number for a book when most are words
	bookRegex := regexp.MustCompile(`(?i)(book|chapter) (\d+)\W`)
	bookmarkTitle = bookRegex.ReplaceAllStringFunc(bookmarkTitle, func(s string) string {
		parts := strings.Split(s, " ")

		nonNumeric := regexp.MustCompile("[^0-9]+")
		numStr := nonNumeric.ReplaceAllString(parts[1], "")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			println(err.Error())
			return bookmarkTitle
		}
		// Put it back, but with an extra colon in there. Some of the `Book X` bookmarks don't have one, and this
		// messes things up. If we put one in we might split twice, but then we remove the empty strings from the
		// array.
		return fmt.Sprintf(":%s %s:", parts[0], num2words.Convert(num))
	})

	// If the string contains Bulletopia, then deal with it as an exception. The bookmarking consistency is atrocious
	if strings.Contains(bookmarkTitle, "Bulletopia") {
		bookmarkTitle = strings.Replace(bookmarkTitle, "Bulletopia", ": Bulletopia :", 1)
	}
	splitRegex := regexp.MustCompile(`([:_"()]|(- )|\.{3})`)
	parts := splitRegex.Split(bookmarkTitle, -1)
	parts = slices.DeleteFunc(parts, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	if len(parts) == 3 {
		// Three-way split? Series, storyline, episodeNumber
		series = CapitalizeWords(strings.TrimSpace(parts[0]))
		storyline = CapitalizeWords(strings.TrimSpace(parts[1]))
		episodeNumber = extractPartNumberFromString(parts[2])

		return
	}

	partFinder := regexp.MustCompile(`(?i)^.*(?P<whole>part (?P<episodeNumber>\w+)).*$`)
	if partFinder.MatchString(bookmarkTitle) {
		namedResults := FindNamedMatches(partFinder, bookmarkTitle)
		partString := namedResults["episodeNumber"]
		maybePart, err := ParseTextNumber(partString)
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
	series = TrimNonAlphaNumeric(CapitalizeWords(series))
	storyline = TrimNonAlphaNumeric(CapitalizeWords(storyline))

	// Occasionally, we have multiple colons. This looks odd.
	if strings.Count(storyline, ":") > 1 {
		storyline = strings.Replace(storyline, ":", " -", 1)
	}

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
		"Indicia",
		"A Year In Thrills",
	}

	for _, s := range seriesToSkip {
		if seriesTitle == s {
			logger.Info(fmt.Sprintf("Skipping series %s", s))
			return false
		}
	}
	for _, s := range pagesToSkip {
		for _, t := range []string{episodeTitle, seriesTitle} {
			if strings.Contains(strings.ToLower(t), strings.ToLower(s)) || levenshtein.DistanceForStrings([]rune(s), []rune(t), levenshtein.DefaultOptions) < 5 {
				logger.V(1).Info(fmt.Sprintf("\"%s\" contains, or is close to, \"%s\"", t, s))
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
		if maybePart, err = ParseTextNumber(toParse); err == nil {
			part = maybePart
		}
	} else {
		part = maybePart
	}
	return
}

func ExtractCreatorsFromCredits(toParse string) (credits api.Credits) {
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
	tokens := strings.Split(CapitalizeWords(strings.Join(input, " ")), "&")

	creators := make([]string, 0)
	for _, v := range tokens {
		creators = append(creators, strings.TrimSpace(v))
	}
	return creators
}
