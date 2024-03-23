package scan

import (
	"cmp"
	"context"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"golang.org/x/exp/maps"
	"slices"
	"strings"
)

type titleCounts struct {
	Title     string
	Count     int
	FirstSeen int
	LastSeen  int
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

	findTypoedSeries(issues, logger, appEnv)
	findTypoedEpisodes(issues, logger)
	findSwappedSeriesEpisodeTitles(issues, logger)
}

func findSwappedSeriesEpisodeTitles(issues *[]api.Issue, logger logr.Logger) {
	suggestions := make([]Suggestion, 0)
	seriesEpisodes, seriesEpisodeTitleCounts := episodesBySeries(issues)
	seriesNames := maps.Keys(seriesEpisodes)
	for _, seriesName := range seriesNames {
		for _, cmpSeries := range seriesNames {
			if seriesName == cmpSeries {
				continue
			}
			for j, episode := range seriesEpisodes[cmpSeries] {
				if cmpSeries == episode.Title {
					continue
				}
				if j > 0 && episode.Title == seriesEpisodes[cmpSeries][j-1].Title {
					continue
				}
				if episode.Title == seriesName {
					rootCounts, _ := seriesEpisodeTitleCounts[seriesName]
					cmpCounts, _ := seriesEpisodeTitleCounts[cmpSeries]

					rcIdx := slices.IndexFunc(rootCounts, func(a *titleCounts) bool {
						return a.Title == cmpSeries
					})
					cmpIdx := slices.IndexFunc(cmpCounts, func(a *titleCounts) bool {
						return a.Title == seriesName
					})

					if rcIdx < 0 || cmpIdx < 0 {
						continue
					}

					if rootCounts[rcIdx].Count > cmpCounts[cmpIdx].Count {
						logger.Info(fmt.Sprintf("Suggest renaming %s - %s to %s - %s", cmpSeries, episode.Title, seriesName, cmpSeries))
						suggestions = append(suggestions, Suggestion{
							From: cmpSeries,
							To:   episode.Title,
						})
					}
				}
			}
		}
	}

	if len(suggestions) > 0 {
		for _, issue := range *issues {
			for _, episode := range issue.Episodes {
				for _, suggestion := range suggestions {
					if episode.Series == suggestion.From && episode.Title == suggestion.To {
						episode.Series = suggestion.To
						episode.Title = suggestion.From
					}
				}
			}
		}
	}
}

func findTypoedSeries(issues *[]api.Issue, logger logr.Logger, appEnv AppEnv) {
	// Look for series titles that are close to others
	allSeries := seriesTitleCounts(issues)
	suggestions := getSuggestions(logger, appEnv.Known, allSeries, SeriesTitle)
	suggestions = append(suggestions, Suggestion{
		"Dexter", "Sinister Dexter", SeriesTitle,
	}, Suggestion{
		"Sinister", "Sinister Dexter", SeriesTitle,
	})
	for _, issue := range *issues {
		for i, e := range issue.Episodes {
			for _, suggestion := range suggestions {
				if e.Series == suggestion.From {
					e.Series = suggestion.To
					issue.Episodes[i] = e
				}
			}
		}
	}
}

func findTypoedEpisodes(issues *[]api.Issue, logger logr.Logger) {
	// Create a map of series -> episodes
	// For each series, create a count mapping of episode titles.
	// Do the comparisons, as with series titles
	seriesEpisodes, seriesEpisodeTitles := episodesBySeries(issues)

	for k, v := range seriesEpisodeTitles {
		seriesSuggestions := getSuggestions(logger, []string{}, v, EpisodeTitle)
		if len(seriesSuggestions) == 0 {
			continue
		}
		episodes := seriesEpisodes[k]
		for _, ep := range episodes {
			for _, s := range seriesSuggestions {
				if ep.Title == s.From {
					logger.Info("Renaming episode", "series", ep.Series, "from", s.From, "to", s.To)
					ep.Title = s.To
				}
			}
		}
	}
}

func episodesBySeries(issues *[]api.Issue) (map[string][]*api.Episode, map[string][]*titleCounts) {
	seriesEpisodes := make(map[string][]*api.Episode, len(*issues))
	seriesEpisodeTitles := make(map[string][]*titleCounts)
	for _, issue := range *issues {
		for _, ep := range issue.Episodes {
			seriesName := ep.Series
			if episodes, ok := seriesEpisodes[seriesName]; ok {
				seriesEpisodes[seriesName] = append(episodes, ep)
				episodeTitleCounts := seriesEpisodeTitles[seriesName]
				if idx, found := slices.BinarySearchFunc(episodeTitleCounts, &titleCounts{Title: ep.Title, Count: 1}, func(e, t *titleCounts) int {
					return cmp.Compare(e.Title, t.Title)
				}); found {
					c := episodeTitleCounts[idx]
					episodeTitleCounts[idx] = &titleCounts{
						Title:     c.Title,
						Count:     c.Count + 1,
						FirstSeen: min(c.FirstSeen, issue.IssueNumber),
						LastSeen:  max(c.LastSeen, issue.IssueNumber),
					}
				} else {
					seriesEpisodeTitles[seriesName] = slices.Insert(seriesEpisodeTitles[seriesName], idx, &titleCounts{
						ep.Title, 1, issue.IssueNumber, issue.IssueNumber,
					})
				}
			} else {
				seriesEpisodes[seriesName] = []*api.Episode{ep}
				seriesEpisodeTitles[seriesName] = []*titleCounts{{
					ep.Title, 1, issue.IssueNumber, issue.IssueNumber,
				}}
			}
		}
	}
	return seriesEpisodes, seriesEpisodeTitles
}

func getSuggestions(logger logr.Logger, knownTitles []string, results []*titleCounts, suggestionType SuggestionType) (suggestions []Suggestion) {
	for _, k := range results {
		targetDistance := getTargetLevenshteinDistance(k.Title)
		for _, l := range results {
			// If they match or the smaller series is a known title
			if k == l || slices.Contains(knownTitles, l.Title) {
				continue
			}
			kTitle := k.Title
			lTitle := l.Title

			if targetDistance < 3 {
				kTitle = strings.ToLower(k.Title)
				lTitle = strings.ToLower(l.Title)
			}
			distance := levenshtein.DistanceForStrings([]rune(kTitle), []rune(lTitle), levenshtein.DefaultOptions)
			if distance > targetDistance || (distance <= targetDistance && l.Count > k.Count) {
				continue
			}
			//Only suggest a change if l's "seen" range is within k's seen range
			if suggestionType == EpisodeTitle {
				if (l.FirstSeen > k.FirstSeen && l.LastSeen < k.LastSeen) || (k.FirstSeen-l.LastSeen <= 2) || (k.LastSeen-l.FirstSeen <= 2) {
					logger.Info("Suggesting an episode title change", "from", l.Title, "to", k.Title)
					suggestions = append(suggestions, Suggestion{
						From: l.Title,
						To:   k.Title,
						Type: suggestionType,
					})
				}
			} else {
				logger.Info("Suggesting a series title change", "from", l.Title, "to", k.Title)
				suggestions = append(suggestions, Suggestion{
					From: l.Title,
					To:   k.Title,
					Type: suggestionType,
				})
			}
		}
	}
	return
}

func seriesTitleCounts(issues *[]api.Issue) []*titleCounts {
	seriesCounts := make([]*titleCounts, 0, len(*issues)/2)
	for _, issue := range *issues {
		for _, episode := range issue.Episodes {
			if idx, found := slices.BinarySearchFunc(
				seriesCounts,
				&titleCounts{episode.Series, 1, 0, 0},
				func(a, b *titleCounts) int {
					return cmp.Compare(a.Title, b.Title)
				}); !found {
				seriesCounts = slices.Insert(seriesCounts, idx, &titleCounts{episode.Series, 1, issue.IssueNumber, issue.IssueNumber})
			} else {
				seriesCounts[idx].Count++
			}
		}
	}

	slices.SortFunc(seriesCounts, func(i, j *titleCounts) int {
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
