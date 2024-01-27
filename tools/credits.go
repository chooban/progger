//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/env"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	parser := argparse.NewParser("pageobjects", "Try to list page objects")

	filename := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to parse"})
	page := parser.Int("p", "page", &argparse.Options{Required: true, Help: "Page to inspect"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	appEnv := env.NewAppEnv()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	credits, err := scan.Credits(appEnv, *filename, *page, *page+5)

	if err != nil {
		appEnv.Log.Error().Err(err).Msg(fmt.Sprintf("Error extracting credits"))
	}
	appEnv.Log.Info().Msg(fmt.Sprintf("Got credits of '%s'", credits))
}
