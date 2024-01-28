package db

import (
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"gorm.io/gorm"
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

func GetSeriesTitleRenameSuggestions(db *gorm.DB, knownTitles []string) []Suggestion {
	seriesRenames := getSeriesTitleCounts(db)

	return getSuggestions(db, knownTitles, seriesRenames, SeriesTitle)
}

func GetEpisodeTitleRenameSuggestions(db *gorm.DB, knownTitles []string) []Suggestion {
	var allSeries []Series
	db.Find(&allSeries)

	var suggestions []Suggestion

	for _, v := range allSeries {
		var episodeCounts []suggestionsResults
		db.Model(&Episode{}).
			Select("title, count(*) as count").
			Where("series_id = ?", v.ID).
			Group("title").
			Find(&episodeCounts)

		if len(episodeCounts) < 2 {
			// Not going to be any renaming if there's one storyline
			continue
		}
		suggestions = append(suggestions, getSuggestions(db, knownTitles, episodeCounts, EpisodeTitle)...)
	}

	return suggestions
}

// ApplySuggestion updates the database in db.DB in line with the
// instructions in the suggestion.
// It should take the suggestion.From value and find a Series object
// with that as a title. Then find a series with suggestion.To as a title.
// All episodes connected to the first should be updated to point to the
// second series. The first series should then be deleted.
func ApplySuggestion(db *gorm.DB, suggestion Suggestion) {
	var fromSeries, toSeries Series
	var episodes []Episode

	//db.Log.Info().Msg(fmt.Sprintf("Moving all episodes linked to '%s' to '%s' instead", suggestion.From, suggestion.To))

	// Find the series with the title suggestion.From
	db.Where("title = ?", suggestion.From).First(&fromSeries)

	// Find the series with the title suggestion.To
	db.Where("title = ?", suggestion.To).First(&toSeries)

	// Update all episodes connected to the first series to point to the second series
	db.Model(&episodes).Where("series_id = ?", fromSeries.ID).Update("series_id", toSeries.ID)

	// Delete the first series
	db.Delete(&fromSeries)
}

func getSeriesTitleCounts(db *gorm.DB) []suggestionsResults {
	var results []suggestionsResults

	db.Raw(
		`select series.title, count(*) as count from series 
			 join episodes on (episodes.series_id = series.id) 
			 group by series.title order by series.title ASC`,
	).Scan(&results)

	return results
}

func getSuggestions(db *gorm.DB, knownTitles []string, results []suggestionsResults, suggestionType SuggestionType) (suggestions []Suggestion) {
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
