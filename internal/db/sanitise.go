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

	for _, k := range results {
		for _, l := range results {
			// If they match or the smaller series is a known title
			if k == l || slices.Contains(appEnv.Known.SeriesTitles, l.Title) {
				continue
			}
			distance := levenshtein.DistanceForStrings([]rune(k.Title), []rune(l.Title), levenshtein.DefaultOptions)
			if distance > 5 || (distance <= 5 && l.Count > k.Count) {
				continue
			}
			var target = k.Title
			var toChange = l.Title

			appEnv.Log.Info().Msgf("Suggest changing '%s' to '%s'", toChange, target)
		}
	}
}
