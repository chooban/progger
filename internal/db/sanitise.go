package db

import (
	"github.com/chooban/progdl-go/internal/env"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"slices"
)

type suggestionsResults struct {
	Title string
	Count int
}

func Suggestions(appEnv env.AppEnv) {
	var results []suggestionsResults

	appEnv.Db.Raw(
		`select series.title, count(*) as count from series 
			 join episodes on (episodes.series_id = series.id) 
			 group by series.title order by series.title ASC`,
	).Scan(&results)

	getSuggestions(appEnv, results)
}

func getSuggestions(appEnv env.AppEnv, results []suggestionsResults) {
	for _, k := range results {
		for _, l := range results {
			// If they match or the smaller series is a known title
			if k == l || slices.Contains(appEnv.Known.SeriesTitles, l.Title) {
				continue
			}
			targetDistance := getTargetLevenshteinDistance(l.Title)
			distance := levenshtein.DistanceForStrings([]rune(k.Title), []rune(l.Title), levenshtein.DefaultOptions)
			if distance > targetDistance || (distance <= targetDistance && l.Count > k.Count) {
				continue
			}
			var target = k.Title
			var toChange = l.Title

			appEnv.Log.Info().Msgf("Suggest changing '%s' to '%s'", toChange, target)
		}
	}
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
