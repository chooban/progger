package app

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func WithLogger() (context.Context, context.CancelFunc, logr.Logger) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(writer)
	logger = logger.With().Caller().Logger()
	var log = zerologr.New(&logger)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = logr.NewContext(ctx, log)

	return ctx, cancel, log
}
