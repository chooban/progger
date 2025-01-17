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

	progs, _ := download.ListAvailableIssues(ctx)

	println(fmt.Sprintf("%+v", progs))
}
