//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/pdfium"
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
	appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	credits, err := appEnv.Pdf.Credits(*filename, *page, *page+5)

	appEnv.Log.Info().Msg(fmt.Sprintf("Got credits of '%s'", credits))

}
