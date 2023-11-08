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

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	issues := scanner.ScanDir(appEnv, *d)

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
