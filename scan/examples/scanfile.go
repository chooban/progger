//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func main() {

	parser := argparse.NewParser("scan", "Scans a pdf")
	f := parser.String("f", "file", &argparse.Options{Required: true, Help: "Directory to scan"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	var log = zerologr.New(&logger)

	ctx := logr.NewContext(context.Background(), log)

	// Create a scanner with no known series or skip titles
	scanner := scan.NewScanner([]string{}, []string{})
	issue, err := scanner.File(ctx, *f)
	if err != nil {
		log.Error(err, "Failed to scan file")
		os.Exit(1)
	}

	for _, v := range issue.Episodes {
		log.Info(fmt.Sprintf("Series: %s", v.Series))
		log.Info(fmt.Sprintf("Title: %s", v.Title))
		log.Info(fmt.Sprintf("Writers: %+v", v.Credits))
	}
}
