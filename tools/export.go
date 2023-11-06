//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/export"
	"os"
)

func main() {
	parser := argparse.NewParser("export", "Exports a PDF of selected series")

	series := parser.String("s", "series", &argparse.Options{Required: true, Help: "Series name"})
	episodeTitle := parser.String("e", "episodes", &argparse.Options{Required: false, Help: "Episode title"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")

	export.BuildPdf(appEnv, *series, *episodeTitle)
}
