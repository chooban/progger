package db

import (
	"fmt"
	"github.com/chooban/progdl-go/internal/env"
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

func GetSeriesTitleRenameSuggestions(appEnv env.AppEnv) []Suggestion {
	seriesRenames := getSeriesTitleCounts(appEnv)

	return getSuggestions(appEnv, seriesRenames, SeriesTitle)
}

func GetEpisodeTitleRenameSuggestions(appEnv env.AppEnv) []Suggestion {
	var allSeries []Series
	appEnv.Db.Find(&allSeries)

	var suggestions []Suggestion

	for _, v := range allSeries {
		var episodeCounts []suggestionsResults
		appEnv.Db.Model(&Episode{}).
			Select("title, count(*) as count").
			Where("series_id = ?", v.ID).
			Group("title").
			Find(&episodeCounts)

		if len(episodeCounts) < 2 {
			// Not going to be any renaming if there's one storyline
			continue
		}
		suggestions = append(suggestions, getSuggestions(appEnv, episodeCounts, EpisodeTitle)...)
	}

	return suggestions
}

// ApplySuggestion updates the database in appEnv.DB in line with the
// instructions in the suggestion.
// It should take the suggestion.From value and find a Series object
// with that as a title. Then find a series with suggestion.To as a title.
// All episodes connected to the first should be updated to point to the
// second series. The first series should then be deleted.
func ApplySuggestion(appEnv env.AppEnv, suggestion Suggestion) {
	var fromSeries, toSeries Series
	var episodes []Episode

	appEnv.Log.Info().Msg(fmt.Sprintf("Moving all episodes linked to '%s' to '%s' instead", suggestion.From, suggestion.To))

	// Find the series with the title suggestion.From
	appEnv.Db.Where("title = ?", suggestion.From).First(&fromSeries)

	// Find the series with the title suggestion.To
	appEnv.Db.Where("title = ?", suggestion.To).First(&toSeries)

	// Update all episodes connected to the first series to point to the second series
	appEnv.Db.Model(&episodes).Where("series_id = ?", fromSeries.ID).Update("series_id", toSeries.ID)

	// Delete the first series
	appEnv.Db.Delete(&fromSeries)
}

func getSeriesTitleCounts(appEnv env.AppEnv) []suggestionsResults {
	var results []suggestionsResults

	appEnv.Db.Raw(
		`select series.title, count(*) as count from series 
			 join episodes on (episodes.series_id = series.id) 
			 group by series.title order by series.title ASC`,
	).Scan(&results)

	return results
}

func getSuggestions(appEnv env.AppEnv, results []suggestionsResults, suggestionType SuggestionType) (suggestions []Suggestion) {
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
			suggestions = append(suggestions, Suggestion{
				From: l.Title,
				To:   k.Title,
				Type: suggestionType,
			})
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
