package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chooban/progger/download"
	"github.com/chooban/progger/download/api"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func main() {
	ctx, logger := withLogger(context.Background())
	ctx = download.WithLoginDetails(ctx, os.Getenv("REBELLION_USERNAME"), os.Getenv("REBELLION_PASSWORD"))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error(err, "Could not determine home dir")
	}

	var printUsage bool
	flag.BoolVar(&printUsage, "help", false, "Print usage")

	var downloadDir string
	flag.StringVar(&downloadDir, "download-dir", homeDir, "Directory for downloads")

	var downloadCount = 5
	flag.IntVar(&downloadCount, "download-count", downloadCount, "Number of progs to download")

	flag.Parse()

	if printUsage {
		flag.PrintDefaults()
		return
	}

	list, err := download.ListAvailableProgs(ctx)

	if err != nil {
		logger.Error(err, "Error listing progs")
	}

	if len(list) == 0 {
		logger.Info("No progs found")
		return
	}

	for i := 0; i < downloadCount; i++ {
		logger.Info(fmt.Sprintf("Downloading %s...", list[i].IssueNumber))
	}
	for i := 0; i <= downloadCount && i < len(list); i++ {
		if filepath, err := download.Download(ctx, list[i], downloadDir, api.Pdf); err != nil {
			logger.Error(err, "could not download file")
		} else {
			logger.Info("Downloaded a file", "file", filepath)
		}
	}
}

func withLogger(ctx context.Context) (context.Context, logr.Logger) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zlogger := zerolog.New(writer)
	zlogger = zlogger.With().Caller().Timestamp().Logger()

	var logrLogger = zerologr.New(&zlogger)
	ctx = logr.NewContext(ctx, logrLogger)

	return ctx, logrLogger
}
