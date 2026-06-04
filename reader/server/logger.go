package server

import (
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func SetupLogger(level string) logr.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logLevel := zerolog.InfoLevel
	if err := logLevel.UnmarshalText([]byte(level)); err == nil {
		zerolog.SetGlobalLevel(logLevel)
	}
	logger := zerolog.New(writer).With().Caller().Logger()
	return zerologr.New(&logger)
}
