package scanner

import (
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
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
		Known: env.ToSkip{SeriesTitles: []string{
			"Strontium Dog",
			"Strontium Dug",
			"The Fall of Deadworld",
		}},
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
			expectedTitle:  "Hunted",
		},
		{
			name:           "Counterfeit Girl",
			input:          "Counterfeit Girl - Part 3",
			expectedPart:   3,
			expectedSeries: "Counterfeit Girl",
			expectedTitle:  "Counterfeit Girl",
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
			expectedTitle:  "Book Ten: The Marze Murderer",
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
			expectedTitle:  "Cover",
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
			expectedTitle:  "Caballistics, Inc.",
		},
		{
			name:           "Tales from MC1",
			input:          "Tales from Mega-City One - Christmas comes to Devil's Island",
			expectedPart:   1,
			expectedSeries: "Tales From Mega-City One",
			expectedTitle:  "Christmas Comes To Devil's Island",
		},
		{
			name:           "Bulletopia",
			input:          "Sinister Dexter Bulletopia: Chapter 2 Stay Brave - Part 1",
			expectedPart:   1,
			expectedSeries: "Sinister Dexter Bulletopia",
			expectedTitle:  "Chapter 2 Stay Brave",
		},
		{
			name:           "Bulletopia 2",
			input:          "Sinister Dexter: Bulletopia - Chapter One: Boys In The Hud",
			expectedPart:   1,
			expectedSeries: "Sinister Dexter",
			expectedTitle:  "Bulletopia: Chapter One: Boys In The Hud",
		},
		{
			name:           "Bulletopia 3",
			input:          "Sinister Dexter- Bulletopia Chapter Three: Ghostlands Part One",
			expectedPart:   1,
			expectedSeries: "Sinister Dexter",
			expectedTitle:  "Bulletopia Chapter Three",
			//expectedTitle:  "Bulletopia Chapter Three: Ghostlands",
		},
		{
			name:           "Hope 1",
			input:          "Hope... In The Shadows - Reel One - Part 10",
			expectedPart:   10,
			expectedSeries: "Hope",
			expectedTitle:  "In The Shadows: Reel One",
		},
		{
			name:           "Ace Trucking",
			input:          "Ace Trucking Co.: The Festive Flip-Flop!",
			expectedPart:   1,
			expectedSeries: "Ace Trucking Co.",
			expectedTitle:  "The Festive Flip-Flop!",
		},
		{
			name:           "Nakka",
			input:          "Tharg's 3rillers Present Nakka of the S.T.A.R.S: Part One",
			expectedPart:   1,
			expectedSeries: "Tharg's 3rillers Present Nakka of The S.t.a.r.s",
			expectedTitle:  "Tharg's 3rillers Present Nakka of The S.t.a.r.s",
		},
		{
			name:           "'Splorers",
			input:          "'Splorers",
			expectedPart:   1,
			expectedSeries: "'Splorers",
			expectedTitle:  "'Splorers",
		},
		{
			name:           "Ampney",
			input:          "Ampney Crucis Investigates... - Setting Son",
			expectedPart:   1,
			expectedSeries: "Ampney Crucis Investigates",
			expectedTitle:  "Setting Son",
		},
		{
			name:           "Full Tilt Boogie",
			input:          "Full Tilt Boogie - Part 1",
			expectedPart:   1,
			expectedSeries: "Full Tilt Boogie",
			expectedTitle:  "Full Tilt Boogie",
		},
		{
			name:           "Brink (part in brackets)",
			input:          "Brink - Mercury Retrograde (Part 12)",
			expectedPart:   12,
			expectedSeries: "Brink",
			expectedTitle:  "Mercury Retrograde",
		},
		{
			name:           "Scarlet Traces: Cold War: Book 2",
			input:          "Scarlet Traces: Cold War: Book 2 - Part 12",
			expectedPart:   12,
			expectedSeries: "Scarlet Traces",
			expectedTitle:  "Cold War: Book Two",
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
		{
			name: "Interrogation",
			input: db.Episode{
				Title:  "Doug Church",
				Series: db.Series{Title: "Interrogation"},
				Part:   1,
			},
		},
	}

	appEnv := createAppEnv()
	appEnv.Skip = env.ToSkip{
		SeriesTitles: []string{"Interrogation"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldIncludeEpisode(appEnv, tc.input)
			if got != tc.shouldInclude {
				t.Errorf("shouldIncludeEpisode(%v) = %v; want %v", tc.input.Series.Title+", "+tc.input.Title, got, tc.shouldInclude)
			}
		})
	}
}

func TestFromRawEpisodes(t *testing.T) {
	testCases := []struct {
		name             string
		rawEpisodes      []RawEpisode
		expectedEpisodes []db.Episode
	}{
		{
			name: "Test Case 1",
			rawEpisodes: []RawEpisode{
				{
					Series:    "Test Series",
					Title:     "Test Title",
					Part:      1,
					FirstPage: 1,
					LastPage:  10,
				},
			},
			expectedEpisodes: []db.Episode{
				{
					Series: db.Series{Title: "Test Series"},
					Title:  "Test Title",
					Part:   1,
				},
			},
		}, {
			name: "Nerve Centre",
			rawEpisodes: []RawEpisode{
				{
					Series:    "Nerve Centre",
					Title:     "Nerve Centre",
					Part:      1,
					FirstPage: 1,
					LastPage:  10,
				},
			},
			expectedEpisodes: []db.Episode{},
		},
		// Add more test cases here
	}

	appEnv := createAppEnv()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issue := fromRawEpisodes(appEnv, tc.rawEpisodes)

			for i, expectedExp := range tc.expectedEpisodes {
				ep := issue[i]
				assert.Equal(t, expectedExp.Series.Title, ep.Series.Title)
				assert.Equal(t, expectedExp.Title, ep.Title)
				assert.Equal(t, expectedExp.Part, ep.Part)
			}
		})
	}
}

func discardingLogger() *zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        io.Discard,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)

	return &logger
}

func TestBuildEpisodes(t *testing.T) {
	testCases := []struct {
		name           string
		bookmarks      []pdfcpu.Bookmark
		expectedSeries string
		expectedTitle  string
		expectedPart   int
	}{
		{
			name: "Test Case 1",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "Test Series: Test Title - Part 1",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedSeries: "Test Series",
			expectedTitle:  "Test Title",
			expectedPart:   1,
		},
		{
			name: "Renaming Deadworld",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "The Fall of Deadwood - Jessica",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedPart:   1,
			expectedTitle:  "Jessica",
			expectedSeries: "The Fall of Deadworld",
		},
		{
			name: "Strontium Dog",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "Strontium Dog - Series Title",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedPart:   1,
			expectedSeries: "Strontium Dog",
			expectedTitle:  "Series Title",
		},
		{
			name: "Strontium Dug",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "Strontium Dug - Series Title",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedPart:   1,
			expectedSeries: "Strontium Dug",
			expectedTitle:  "Series Title",
		},
		{
			name: "ABC Warriors",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "Abc Warriors - Series Title",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedPart:   1,
			expectedSeries: "ABC Warriors",
			expectedTitle:  "Series Title",
		},
		{
			name: "The ABC Warriors",
			bookmarks: []pdfcpu.Bookmark{
				{
					Title:    "The Abc Warriors - Series Title",
					PageFrom: 1,
					PageThru: 10,
				},
			},
			expectedPart:   1,
			expectedSeries: "ABC Warriors",
			expectedTitle:  "Series Title",
		},
		// Add more test cases here
	}

	appEnv := env.AppEnv{
		Log: discardingLogger(),
		Db:  nil,
		Known: env.ToSkip{SeriesTitles: []string{
			"ABC Warriors",
			"Strontium Dog",
			"Strontium Dug",
			"The Fall of Deadworld",
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issue := buildIssue(appEnv, "2000AD 123 (1977).pdf", tc.bookmarks)
			assert.Equal(t, 123, issue.IssueNumber)
			assert.Equal(t, tc.expectedSeries, issue.Episodes[0].Series.Title)
			assert.Equal(t, tc.expectedTitle, issue.Episodes[0].Title)
			assert.Equal(t, tc.expectedPart, issue.Episodes[0].Part)
		})
	}
}
