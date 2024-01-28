//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/internal/db"
	"github.com/chooban/progger/scan/types"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"time"

	"github.com/chooban/progger/scan"
	"os"
)

func main() {

	parser := argparse.NewParser("scan", "Scans a directory for progs")
	d := parser.String("d", "directory", &argparse.Options{Required: true, Help: "Directory to scan"})
	c := parser.Int("c", "count", &argparse.Options{Required: false, Help: "Number of issues to scan"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(writer)
	logger = logger.With().Caller().Timestamp().Logger()
	var log = zerologr.New(&logger)

	//appEnv := env.NewAppEnv()
	var myDb = db.Init("progs.db")

	ctx := logr.NewContext(context.Background(), log)

	scan.Dir(ctx, *d, *c)

	//dbIssues := fromRawEpisodes(appEnv, issues[0].Episodes)
	//db.SaveIssues(myDb, issues)
	knownTitles := []string{
		"Anderson, Psi-Division",
		"Strontium Dug",
	}

	suggestions := db.GetSeriesTitleRenameSuggestions(myDb, knownTitles)

	for _, s := range suggestions {
		db.ApplySuggestion(myDb, s)
	}

	suggestions = db.GetEpisodeTitleRenameSuggestions(myDb, knownTitles)

	for _, v := range suggestions {
		log.Info(fmt.Sprintf("Suggest renaming '%s' to '%s'", v.From, v.To))
	}
}

func creators(names []string) (creators []*db.Creator) {
	creators = make([]*db.Creator, len(names))
	for i, v := range names {
		creators[i] = &db.Creator{Name: v}
	}
	return
}

func fromRawEpisodes(rawEpisodes []types.Episode) []db.Episode {
	episodes := make([]db.Episode, 0, len(rawEpisodes))
	for _, rawEpisode := range rawEpisodes {
		writers := creators(rawEpisode.Credits[types.Script])
		artists := creators(rawEpisode.Credits[types.Art])
		colourists := creators(rawEpisode.Credits[types.Colours])
		letterists := creators(rawEpisode.Credits[types.Letters])

		ep := db.Episode{
			Title:    rawEpisode.Title,
			Part:     rawEpisode.Part,
			Series:   db.Series{Title: rawEpisode.Series},
			PageFrom: rawEpisode.FirstPage,
			PageThru: rawEpisode.LastPage,
			Script:   writers,
			Art:      artists,
			Colours:  colourists,
			Letters:  letterists,
		}
		episodes = append(episodes, ep)
	}
	return episodes
}
