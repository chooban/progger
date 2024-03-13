package exporter

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func WithLogger() context.Context {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(writer)
	logger = logger.With().Caller().Timestamp().Logger()
	var log = zerologr.New(&logger)

	ctx := context.Background()
	ctx = logr.NewContext(ctx, log)

	return ctx
}
