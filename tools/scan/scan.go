//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/scanner"
	"github.com/rs/zerolog"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
)

func main() {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(writer)

	parser := argparse.NewParser("scan", "Scans a directory for progs")
	d := parser.String("d", "directory", &argparse.Options{Required: true, Help: "Directory to scan"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	if err := os.Remove("progs.db"); err != nil {
		panic(err)
	}

	appEnv := env.AppEnv{
		Db:  db.Init("progs.db"),
		Log: &logger,
		Skip: env.ToSkip{
			SeriesTitles: []string{
				"Interrogation",
				"New Books",
				"Obituary",
				"Tribute",
				"Untitled",
			},
		},
		Known: env.ToSkip{
			SeriesTitles: []string{
				"Anderson, Psi-Division",
				"Strontium Dug",
			},
		},
	}
	issues := scanner.ScanDir(appEnv, *d)

	db.SaveIssues(appEnv, issues)

	suggestions := db.GetSeriesTitleRenameSuggestions(appEnv)

	for _, s := range suggestions {
		db.ApplySuggestion(appEnv, s)
	}

	suggestions = db.GetEpisodeTitleRenameSuggestions(appEnv)

	for _, v := range suggestions {
		appEnv.Log.Info().Msg(fmt.Sprintf("Suggest renaming '%s' to '%s'", v.From, v.To))
	}
}
