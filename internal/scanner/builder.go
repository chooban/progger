package scanner

import (
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/stringutils"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/rs/zerolog"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"regexp"
	"strconv"
	"strings"
)

func buildEpisodes(appEnv env.AppEnv, issueNumber int, bookmarks []pdfcpu.Bookmark) db.Issue {
	log := appEnv.Log
	allEpisodes := make([]RawEpisode, 0)
	for _, b := range bookmarks {
		part, series, title := extractDetailsFromPdfBookmark(b.Title)

		if series == "" {
			log.Debug().Msg(fmt.Sprintf("Odd title: %s", b.Title))
			continue
		}
		allEpisodes = append(allEpisodes, RawEpisode{
			Title:     title,
			Series:    series,
			Part:      part,
			FirstPage: b.PageFrom,
			LastPage:  b.PageThru,
		})
	}
	issue := db.Issue{
		Publication: db.Publication{Title: "2000 AD"},
		IssueNumber: issueNumber,
	}
	issue.Episodes = fromRawEpisodes(appEnv.Log, allEpisodes)

	return issue
}

func extractDetailsFromPdfBookmark(bookmarkTitle string) (episodeNumber int, series string, storyline string) {
	// We don't want any zero parts. It's 1 if not specified
	episodeNumber = -1
	f := func(c rune) bool {
		return c == ':' || c == '-' || c == '_' || c == '"'
	}
	// We're going to re-capitalise it later, so lowercase it now
	bookmarkTitle = strings.ToLower(bookmarkTitle)

	if len(strings.FieldsFunc(bookmarkTitle, f)) == 3 {
		// Three episodeNumber split? Series, storyline, episodeNumber
		parts := strings.FieldsFunc(bookmarkTitle, f)
		series = stringutils.CapitalizeWords(strings.TrimSpace(parts[0]))
		storyline = stringutils.CapitalizeWords(strings.TrimSpace(parts[1]))
		episodeNumber = extractPartNumberFromString(parts[2])

		return
	}

	partFinder := regexp.MustCompile("^.*(?P<whole>part (?P<episodeNumber>\\w+)).*$")
	if partFinder.MatchString(bookmarkTitle) {
		namedResults := stringutils.FindNamedMatches(partFinder, bookmarkTitle)
		partString := namedResults["episodeNumber"]
		maybePart, err := stringutils.ParseTextNumber(partString)
		if err == nil {
			episodeNumber = maybePart
		}
		toReplace := regexp.MustCompile("\\s+part " + partString + "[^a-zA-Z0-9]*")
		bookmarkTitle = toReplace.ReplaceAllString(bookmarkTitle, " ")
	}
	multiSplit := func(r rune) bool {
		return r == ':' || r == '-'
	}
	titleSplit := strings.FieldsFunc(bookmarkTitle, multiSplit)
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

func fromRawEpisodes(log *zerolog.Logger, rawEpisodes []RawEpisode) []db.Episode {
	episodes := make([]db.Episode, 0, len(rawEpisodes))
	for _, rawEpisode := range rawEpisodes {
		ep := db.Episode{
			Title:  rawEpisode.Title,
			Part:   rawEpisode.Part,
			Series: db.Series{Title: rawEpisode.Series},
		}
		if shouldIncludeEpisode(log, ep) {
			episodes = append(episodes, ep)
		} else {
			log.Info().Msg(fmt.Sprintf("Skipping. Series: %s. Episode: %s", ep.Series.Title, ep.Title))
		}
	}
	return episodes
}

func shouldIncludeEpisode(log *zerolog.Logger, episode db.Episode) bool {
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
	}
	for _, s := range pagesToSkip {
		for _, t := range []string{episode.Title, episode.Series.Title} {
			if stringutils.ContainsI(t, s) || levenshtein.DistanceForStrings([]rune(s), []rune(t), levenshtein.DefaultOptions) < 5 {
				log.Info().Msg(fmt.Sprintf("%s contains, or is close to, %s", t, s))
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
