//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
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

	if err := os.Remove("progs.db"); err != nil {
		panic(err)
	}

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")
	scanner.ScanDir(appEnv, *d)

	//db.SaveIssues(appEnv, issues)
	//
	//suggestions := db.GetSeriesTitleRenameSuggestions(appEnv)
	//
	//for _, s := range suggestions {
	//	db.ApplySuggestion(appEnv, s)
	//}
	//
	//suggestions = db.GetEpisodeTitleRenameSuggestions(appEnv)
	//
	//for _, v := range suggestions {
	//	appEnv.Log.Info().Msg(fmt.Sprintf("Suggest renaming '%s' to '%s'", v.From, v.To))
	//}
}
