package scan

import (
	"cmp"
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
	allSeries := seriesTitleCounts(issues)
	suggestions := getSuggestions(appEnv.Known, allSeries, SeriesTitle)
	applySuggestions(logger, suggestions, issues)

	// Get all the series titles again
	//allSeries = seriesTitleCounts(issues)
	//titleCounts := make(map[string]map[string]suggestionsResults)
	//for _, issue := range *issues {
	//	for _, e := range issue.Episodes {
	//		if _, ok := titleCounts[e.Title]; !ok {
	//			titleCounts[e.Series] = make(map[string]suggestionsResults)
	//		}
	//		if _, ok := titleCounts[e.Series][e.Title]; !ok {
	//			titleCounts[e.Series][e.Title] = suggestionsResults{e.Title, 1}
	//		} else {
	//			titleCounts[e.Series][e.Title] = suggestionsResults{
	//				e.Title, titleCounts[e.Series][e.Title].Count + 1,
	//			}
	//		}
	//	}
	//}

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

func getSuggestions(knownTitles []string, results []*suggestionsResults, suggestionType SuggestionType) (suggestions []Suggestion) {
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

func seriesTitleCounts(issues *[]api.Issue) []*suggestionsResults {
	seriesCounts := make([]*suggestionsResults, 0, len(*issues)/2)
	for _, issue := range *issues {
		for _, episode := range issue.Episodes {
			if idx, found := slices.BinarySearchFunc(
				seriesCounts,
				&suggestionsResults{episode.Series, 1},
				func(a, b *suggestionsResults) int {
					return cmp.Compare(a.Title, b.Title)
				}); !found {
				fmt.Printf("Didn't find anything for %s\n", episode.Series)
				seriesCounts = slices.Insert(seriesCounts, idx, &suggestionsResults{episode.Series, 1})
			} else {
				seriesCounts[idx].Count++
			}
		}
	}

	slices.SortFunc(seriesCounts, func(i, j *suggestionsResults) int {
		return cmp.Compare(len(i.Title), len(j.Title))
	})

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
