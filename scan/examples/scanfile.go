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
	issue, _ := scan.File(ctx, *f)

	for _, v := range issue.Episodes {
		log.Info(fmt.Sprintf("Series: %s", v.Series))
		log.Info(fmt.Sprintf("Title: %s", v.Title))
		log.Info(fmt.Sprintf("Writers: %+v", v.Credits))
	}
}
