package scanner

import (
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/stringutils"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func buildIssue(appEnv env.AppEnv, filename string, bookmarks []pdfcpu.Bookmark) db.Issue {
	log := appEnv.Log
	issueNumber, _ := getProgNumber(filename)
	allEpisodes := make([]RawEpisode, 0)
	for _, b := range bookmarks {
		part, series, title := extractDetailsFromPdfBookmark(b.Title)

		if series == "" {
			log.Debug().Msg(fmt.Sprintf("Odd title: %s", b.Title))
			continue
		}
		// Check to see if the series is close to any of the blessed titles
		for _, v := range appEnv.Known.SeriesTitles {
			if series == v {
				break
			}
			distance := levenshtein.DistanceForStrings(
				[]rune(strings.ToLower(v)),
				[]rune(strings.ToLower(series)),
				levenshtein.DefaultOptions,
			)
			log.Debug().Msg(fmt.Sprintf("Distance between '%s' and '%s' is %d", v, series, distance))
			if distance < 5 {
				series = v
			}
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
		Filename:    filepath.Base(filename),
	}
	issue.Episodes = fromRawEpisodes(appEnv, allEpisodes)

	return issue
}

func extractDetailsFromPdfBookmark(bookmarkTitle string) (episodeNumber int, series string, storyline string) {
	// We don't want any zero parts. It's 1 if not specified
	episodeNumber = -1
	bookmarkTitle = strings.ToLower(bookmarkTitle)

	splitRegex := regexp.MustCompile("([:_\"]|(- ))")
	parts := splitRegex.Split(bookmarkTitle, -1)

	if len(parts) == 3 {
		// Three episodeNumber split? Series, storyline, episodeNumber
		series = stringutils.CapitalizeWords(strings.TrimSpace(parts[0]))
		storyline = stringutils.CapitalizeWords(strings.TrimSpace(parts[1]))
		episodeNumber = extractPartNumberFromString(parts[2])

		return
	}

	partFinder := regexp.MustCompile(`^.*(?P<whole>part (?P<episodeNumber>\w+)).*$`)
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

	titleSplit := splitRegex.Split(bookmarkTitle, -1)
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

func fromRawEpisodes(appEnv env.AppEnv, rawEpisodes []RawEpisode) []db.Episode {
	episodes := make([]db.Episode, 0, len(rawEpisodes))
	for _, rawEpisode := range rawEpisodes {
		ep := db.Episode{
			Title:    rawEpisode.Title,
			Part:     rawEpisode.Part,
			Series:   db.Series{Title: rawEpisode.Series},
			PageFrom: rawEpisode.FirstPage,
			PageThru: rawEpisode.LastPage,
		}
		if shouldIncludeEpisode(appEnv, ep) {
			episodes = append(episodes, ep)
		} else {
			appEnv.Log.Info().Msg(fmt.Sprintf("Skipping. Series: %s. Episode: %s", ep.Series.Title, ep.Title))
		}
	}
	return episodes
}

func shouldIncludeEpisode(appEnv env.AppEnv, episode db.Episode) bool {
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
	log := appEnv.Log

	for _, s := range appEnv.Skip.SeriesTitles {
		if episode.Series.Title == s {
			log.Info().Msg(fmt.Sprintf("Skipping series %s", s))
			return false
		}
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
