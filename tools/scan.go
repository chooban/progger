//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/pdfium"
	"github.com/chooban/progdl-go/internal/scanner"
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

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	issues := scanner.ScanDir(appEnv, *d, *c)

	db.SaveIssues(appEnv.Db, issues)

	suggestions := db.GetSeriesTitleRenameSuggestions(appEnv.Db, appEnv.Known.SeriesTitles)

	for _, s := range suggestions {
		db.ApplySuggestion(appEnv.Db, s)
	}

	suggestions = db.GetEpisodeTitleRenameSuggestions(appEnv.Db, appEnv.Known.SeriesTitles)

	for _, v := range suggestions {
		appEnv.Log.Info().Msg(fmt.Sprintf("Suggest renaming '%s' to '%s'", v.From, v.To))
	}
}

func creators(names []string) (creators []*db.Creator) {
	creators = make([]*db.Creator, len(names))
	for i, v := range names {
		creators[i] = &db.Creator{Name: v}
	}
	return
}

func fromRawEpisodes(appEnv env.AppEnv, rawEpisodes []scanner.Episode) []db.Episode {
	episodes := make([]db.Episode, 0, len(rawEpisodes))
	for _, rawEpisode := range rawEpisodes {
		writers := creators(rawEpisode.Credits[scanner.Script])
		artists := creators(rawEpisode.Credits[scanner.Art])
		colourists := creators(rawEpisode.Credits[scanner.Colours])
		letterists := creators(rawEpisode.Credits[scanner.Letters])

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
