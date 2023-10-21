package scanner

import (
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func createAppEnv() env.AppEnv {
	writer := zerolog.ConsoleWriter{
		Out:        io.Discard,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	appEnv := env.AppEnv{
		Log: &logger,
		Db:  nil,
	}
	return appEnv
}

func TestExtractDetailsFromTitle(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedPart   int
		expectedSeries string
		expectedTitle  string
	}{
		{
			name:           "JD1",
			input:          "Judge Dredd: Get Sin - Part 2",
			expectedPart:   2,
			expectedSeries: "Judge Dredd",
			expectedTitle:  "Get Sin",
		},
		{
			name:           "Hunted",
			input:          "Hunted - Part 4",
			expectedPart:   4,
			expectedSeries: "Hunted",
			expectedTitle:  "",
		},
		{
			name:           "Counterfeit Girl",
			input:          "Counterfeit Girl - Part 3",
			expectedPart:   3,
			expectedSeries: "Counterfeit Girl",
			expectedTitle:  "",
		},
		{
			name:           "Nemesis",
			input:          "Nemesis: Tubular Hells",
			expectedPart:   1,
			expectedSeries: "Nemesis",
			expectedTitle:  "Tubular Hells",
		},
		{
			name:           "Savage",
			input:          "Savage: Book 10: The Marze Murderer - Part 2",
			expectedPart:   2,
			expectedSeries: "Savage",
			expectedTitle:  "Book 10: The Marze Murderer",
		},
		{
			name:           "A Dredd",
			input:          "Judge Dredd - The Last Temptation of Joe",
			expectedPart:   1,
			expectedSeries: "Judge Dredd",
			expectedTitle:  "The Last Temptation of Joe",
		},
		{
			name:           "Out 3",
			input:          "The Out - Book Three- Part Three",
			expectedPart:   3,
			expectedSeries: "The Out",
			expectedTitle:  "Book Three",
		},
		{
			name:           "Joe Pineapples",
			input:          "Joe Pineapples - Tin Man - Six",
			expectedPart:   6,
			expectedSeries: "Joe Pineapples",
			expectedTitle:  "Tin Man",
		},
		{
			name:           "Enemy Earth",
			input:          "Enemy Earth - Book One - Part Two",
			expectedPart:   2,
			expectedSeries: "Enemy Earth",
			expectedTitle:  "Book One",
		},
		{
			name:           "Enemy Earth 2",
			input:          "Enemy Earth- Book One: Part Four",
			expectedPart:   4,
			expectedSeries: "Enemy Earth",
			expectedTitle:  "Book One",
		},
		{
			name:           "Dredd - Buratino",
			input:          "Judge Dredd - Buratino Must Die: 04",
			expectedPart:   4,
			expectedSeries: "Judge Dredd",
			expectedTitle:  "Buratino Must Die",
		},
		{
			name:           "Pandora Perfect",
			input:          "Pandora Perfect \"Mystery Moon\" Part Four",
			expectedPart:   4,
			expectedTitle:  "Mystery Moon",
			expectedSeries: "Pandora Perfect",
		},
		{
			name:           "Deadworld 1",
			input:          "The Fall of Deadworld - Damned - part 12",
			expectedPart:   12,
			expectedTitle:  "Damned",
			expectedSeries: "The Fall of Deadworld",
		},
		{
			name:           "Deadworld 2",
			input:          "The Fall Of Deadworld - Damned - part 12",
			expectedPart:   12,
			expectedTitle:  "Damned",
			expectedSeries: "The Fall of Deadworld",
		},
		{
			name:           "3rillers",
			input:          "Tharg's 3rillers Presents: Saphir- Un Roman Fantastique: Part one",
			expectedPart:   1,
			expectedTitle:  "Saphir: Un Roman Fantastique",
			expectedSeries: "Tharg's 3rillers Presents",
		},
		{
			name:           "Hershey",
			input:          "Hershey: The Cold In The Bones - Book One - Part 2",
			expectedPart:   2,
			expectedSeries: "Hershey",
			expectedTitle:  "The Cold In The Bones: Book One",
		},
		{
			name:           "Hershey - Bones",
			input:          "Hershey: Part One - The Cold in the Bones - Book One",
			expectedPart:   1,
			expectedTitle:  "The Cold In The Bones: Book One",
			expectedSeries: "Hershey",
		},
		{
			name:           "Hershey - Bones 2",
			input:          "Hershey - The Cold In The Bones: Book One - Part 7",
			expectedPart:   7,
			expectedSeries: "Hershey",
			expectedTitle:  "The Cold In The Bones: Book One",
		},
		{
			name:           "Cover",
			input:          "Cover",
			expectedPart:   1,
			expectedSeries: "Cover",
			expectedTitle:  "",
		},
		{
			name:           "Skip Tracer",
			input:          "Skip Tracer: Nimrod - Part 4",
			expectedPart:   4,
			expectedSeries: "Skip Tracer",
			expectedTitle:  "Nimrod",
		},
		{
			name:           "Feature",
			input:          "Feature: Caballistics, INC.",
			expectedPart:   1,
			expectedSeries: "Feature",
			expectedTitle:  "Caballistics, Inc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotPart, gotSeries, gotTitle := extractDetailsFromPdfBookmark(tc.input)
			if gotPart != tc.expectedPart || gotSeries != tc.expectedSeries || gotTitle != tc.expectedTitle {
				t.Errorf("extractDetailsFromPdfBookmark(%v) = %v, %v, %v; want %v, %v, %v", tc.input, gotPart, gotSeries, gotTitle, tc.expectedPart, tc.expectedSeries, tc.expectedTitle)
			}
		})
	}
}

func TestShouldIncludeEpisode(t *testing.T) {
	testCases := []struct {
		name          string
		input         db.Episode
		shouldInclude bool
	}{
		{
			name:          "Cover",
			input:         db.Episode{Title: "Cover"},
			shouldInclude: false,
		},
		{
			name:          "Nerve Centre",
			input:         db.Episode{Title: "Nerve Centre"},
			shouldInclude: false,
		},
		{
			name:          "Nerve Center",
			input:         db.Episode{Title: "Nerve Center"},
			shouldInclude: false,
		},
		{
			name:          "Input",
			input:         db.Episode{Title: "Input"},
			shouldInclude: false,
		},
		{
			name:          "Art stars",
			input:         db.Episode{Title: "2000AD Art stars winner"},
			shouldInclude: false,
		},
		{
			name: "Joko's Nerve Centre",
			input: db.Episode{
				Title: "",
				Series: db.Series{
					Title: "Joko-jargo's Nerve Centre",
				},
			},
			shouldInclude: false,
		},
		{
			name: "Alan Grant Pin up",
			input: db.Episode{
				Title: "Alan Grant Pin up",
				Series: db.Series{
					Title: "Alan Grant Pin up",
				},
			},
			shouldInclude: false,
		},
		{
			name: "Dredd Pin up",
			input: db.Episode{
				Title: "Dredd Pin-up",
				Series: db.Series{
					Title: "Dredd Pin-up",
				},
			},
			shouldInclude: false,
		},
		{
			name:          "Regular Episode",
			input:         db.Episode{Title: "Regular Episode"},
			shouldInclude: true,
		},
		{
			name:          "Cover in name",
			input:         db.Episode{Title: "The Radyar Recovery"},
			shouldInclude: true,
		},
		{
			name: "Skip tracer",
			input: db.Episode{
				Title:  "Nimrod",
				Series: db.Series{Title: "Skip Tracer"},
				Part:   4,
			},
			shouldInclude: true,
		},
		{
			name: "Feature",
			input: db.Episode{
				Title:  "Caballistics, Inc",
				Series: db.Series{Title: "Feature"},
				Part:   1,
			},
			shouldInclude: false,
		},
	}

	writer := zerolog.ConsoleWriter{
		Out:        io.Discard,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldIncludeEpisode(&logger, tc.input)
			if got != tc.shouldInclude {
				t.Errorf("shouldIncludeEpisode(%v) = %v; want %v", tc.input.Series.Title+", "+tc.input.Title, got, tc.shouldInclude)
			}
		})
	}
}
func TestFromRawEpisodes(t *testing.T) {
	// Create a mock AppEnv
	appEnv := createAppEnv()

	// Create a mock RawEpisode
	rawEpisodes := []RawEpisode{
		{
			Series:    "Test Series",
			Title:     "Test Title",
			Part:      1,
			FirstPage: 1,
			LastPage:  10,
		},
	}

	issue := fromRawEpisodes(appEnv.Log, rawEpisodes)

	ep := issue[0]
	assert.Equal(t, "Test Series", ep.Series.Title)
	assert.Equal(t, "Test Title", ep.Title)
	assert.Equal(t, 1, ep.Part)
}
