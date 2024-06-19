package main

import (
	"context"
	"fmt"
	"github.com/chooban/progger/download"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func main() {
	ctx, logger := withLogger(context.Background())
	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Error(err, "Could not determine config dir")
		return
	}
	ctx = download.WithLoginDetails(ctx, os.Getenv("REBELLION_USERNAME"), os.Getenv("REBELLION_PASSWORD"))
	ctx = download.WithBrowserContextDir(ctx, configDir)

	start := time.Now()
	list, err := download.ListAvailableProgs(ctx)

	logger.Info(fmt.Sprintf("Found %d progs", len(list)), "duration", time.Since(start))
	logger.Info(fmt.Sprintf("%+v", list[0]))
}

func withLogger(ctx context.Context) (context.Context, logr.Logger) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zlogger := zerolog.New(writer)
	zlogger = zlogger.With().Caller().Logger()

	var logrLogger = zerologr.New(&zlogger)
	ctx = logr.NewContext(ctx, logrLogger)

	return ctx, logrLogger
}
