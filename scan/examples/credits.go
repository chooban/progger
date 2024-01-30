//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"time"
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
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(writer)
	logger = logger.With().Caller().Timestamp().Logger()
	var log = zerologr.New(&logger)

	ctx := context.Background()
	ctx = logr.NewContext(ctx, log)

	credits, err := scan.ReadCredits(ctx, *filename, *page, *page+5)

	if err != nil {
		log.Error(err, fmt.Sprintf("Error extracting credits"))
	}
	log.Info(fmt.Sprintf("Got credits of '%s'", credits))
}
