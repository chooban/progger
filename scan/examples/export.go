//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func main() {
	parser := argparse.NewParser("scan", "Scans a pdf")
	file := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to scan"})
	pageFrom := parser.Int("s", "start", &argparse.Options{Required: true, Help: "Page to export from"})
	pageTo := parser.Int("e", "end", &argparse.Options{Required: true, Help: "Page to export to"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	var log = zerologr.New(&logger)

	ctx := logr.NewContext(context.Background(), log)
	//issue, _ := scan.File(ctx, *file)
	pages := []api.ExportPage{
		{
			Filename: *file,
			PageFrom: *pageFrom,
			PageTo:   *pageTo,
			Title:    "An Example Title",
		},
	}

	err := scan.Build(ctx, pages, false, "export.pdf")
	if err != nil {
		log.Error(err, "Failed to export")
	}

}
