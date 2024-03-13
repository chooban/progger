package scan

import (
	"context"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"slices"
)

type suggestionsResults struct {
	Title string
	Count int
}

type SuggestionType int64

const (
	SeriesTitle SuggestionType = iota
	EpisodeTitle
)

type Suggestion struct {
	From string
	To   string
	Type SuggestionType
}

func Sanitise(ctx context.Context, issues *[]api.Issue) {
	appEnv := fromContextOrDefaults(ctx)
	logger := logr.FromContextOrDiscard(ctx)

	// Look for series titles that are close to others
	allSeries := getAllSeries(issues)
	suggestions := getSuggestions(appEnv.Known, allSeries, SeriesTitle)

	for _, suggestion := range suggestions {
		logger.Info(fmt.Sprintf("%+v", suggestion))
	}
	applySuggestions(logger, suggestions, issues)
}

func applySuggestions(logger logr.Logger, suggestions []Suggestion, issues *[]api.Issue) {
	for _, issue := range *issues {
		for i, e := range issue.Episodes {
			for _, suggestion := range suggestions {
				if e.Series == suggestion.From {
					logger.Info(fmt.Sprintf("%+v", e))
					e.Series = suggestion.To
					issue.Episodes[i] = e
				}
			}
		}
	}
}

func getSuggestions(knownTitles []string, results []suggestionsResults, suggestionType SuggestionType) (suggestions []Suggestion) {
	for _, k := range results {
		for _, l := range results {
			// If they match or the smaller series is a known title
			if k == l || slices.Contains(knownTitles, l.Title) {
				continue
			}
			targetDistance := getTargetLevenshteinDistance(l.Title)
			distance := levenshtein.DistanceForStrings([]rune(k.Title), []rune(l.Title), levenshtein.DefaultOptions)
			if distance > targetDistance || (distance <= targetDistance && l.Count > k.Count) {
				continue
			}
			suggestions = append(suggestions, Suggestion{
				From: l.Title,
				To:   k.Title,
				Type: suggestionType,
			})
		}
	}
	return
}

func getAllSeries(issues *[]api.Issue) []suggestionsResults {
	allSeriesMap := make(map[string]int, len(*issues))
	for _, issue := range *issues {
		for _, episode := range issue.Episodes {
			if _, ok := allSeriesMap[episode.Series]; !ok {
				allSeriesMap[episode.Series] = 1
			} else {
				allSeriesMap[episode.Series]++
			}
		}
	}
	seriesCounts := make([]suggestionsResults, 0, len(allSeriesMap))
	for k, v := range allSeriesMap {
		seriesCounts = append(seriesCounts, suggestionsResults{
			k, v,
		})
	}

	return seriesCounts
}

// getTargetLevenshteinDistance returns the maximum Levenshtein distance
// allowed between two titles when attempting to sanitise them.
// The only parameter is the input string that you might want to change.
// A string of 5 characters or fewer will return a score of 1, scaling
// up to a maximum of 5 if the string is longer than 12 characters.
func getTargetLevenshteinDistance(input string) int {
	length := len(input)
	switch {
	case length <= 5:
		return 1
	case length <= 8:
		return 2
	case length <= 10:
		return 3
	case length <= 12:
		return 4
	default:
		return 5
	}
}
