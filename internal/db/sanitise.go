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

type Suggestion struct {
	From string
	To   string
}

func GetSuggestions(appEnv env.AppEnv) {
	var results []suggestionsResults

	appEnv.Db.Raw(
		`select series.title, count(*) as count from series 
			 join episodes on (episodes.series_id = series.id) 
			 group by series.title order by series.title ASC`,
	).Scan(&results)

	getSuggestions(appEnv, results)
}

// ApplySuggestion updates the database in appEnv.DB in line with the
// instructions in the suggestion.
// It should take the suggestion.From value and find a Series object
// with that as a title. Then find a series with suggestion.To as a title.
// All episodes connectioned to the first should be updated to point to the
// second series. The first series should then be deleted.
func ApplySuggestion(appEnv env.AppEnv, suggestion Suggestion) {
    var fromSeries, toSeries db.Series
    var episodes []db.Episode

    // Find the series with the title suggestion.From
    appEnv.Db.Where("title = ?", suggestion.From).First(&fromSeries)

    // Find the series with the title suggestion.To
    appEnv.Db.Where("title = ?", suggestion.To).First(&toSeries)

    // Update all episodes connected to the first series to point to the second series
    appEnv.Db.Model(&episodes).Where("series_id = ?", fromSeries.ID).Update("series_id", toSeries.ID)

    // Delete the first series
    appEnv.Db.Delete(&fromSeries)
}

func getSuggestions(appEnv env.AppEnv, results []suggestionsResults) (suggestions []Suggestion) {
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
			suggestions = append(suggestions, Suggestion{From: l.Title, To: k.Title})
		}
	}
	return
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
