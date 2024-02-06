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
	ctx = download.WithLoginDetails(ctx, os.Getenv("REBELLION_USERNAME"), os.Getenv("REBELLION_PASSWORD"))

	list, err := download.ListAvailableProgs(ctx)

	if err != nil {
		logger.Error(err, "Error listing progs")
	}

	logger.Info("Successfully reached the end")
	logger.Info(fmt.Sprintf("%+v", list))
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
