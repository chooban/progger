package main

import (
	"context"

	"github.com/chooban/progger/reader"
	"github.com/chooban/progger/reader/server"
)

func main() {
	cfg := server.Load()
	logger := server.SetupLogger(cfg.LogLevel)

	readerCfg := reader.ServerConfig{
		DatabasePath:    cfg.DatabasePath,
		Host:            cfg.Host,
		LibraryName:     cfg.LibraryName,
		ScanDirectories: cfg.ScanDirectories,
		Logger:          logger,
	}

	_, err := reader.Start(context.Background(), readerCfg)
	if err != nil {
		logger.Error(err, "Failed to start")
		return
	}

	select {}
}
