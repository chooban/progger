package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/chooban/progger/reader/config"
	"github.com/chooban/progger/reader/server"
	"github.com/chooban/progger/reader/services"
	"github.com/chooban/progger/scan"
	"github.com/go-logr/logr"
)

func main() {
	cfg := config.Load()
	pdfApi.DisableConfigDir()

	logger := config.SetupLogger(cfg.LogLevel)
	ctx := logr.NewContext(context.Background(), logger)

	db, err := services.OpenDatabase(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error(err, "Failed to open database")
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		logger.Error(err, "Failed to migrate database")
		os.Exit(1)
	}

	// Initialize TSID generator for ID generation
	tsidGen, err := services.NewTSIDGenerator()
	if err != nil {
		logger.Error(err, "Failed to initialize TSID generator")
		os.Exit(1)
	}

	librarySer := services.NewLibraryService(db, tsidGen)
	seriesSer := services.NewSeriesService(db, tsidGen)
	bookSer := services.NewBookService(db, tsidGen)
	coverSer := services.NewCoverService(db, tsidGen)
	pageSer := services.NewPageService(bookSer, nil, scan.BuildPageAsPDF, scan.BuildPageAsImage)
	thumbnailSer := services.NewThumbnailService(pageSer)
	realScanner := services.NewRealScanner(config.KnownSeries, config.SkipTitles)
	scanSer := services.NewScanService(db, cfg, librarySer, seriesSer, bookSer, coverSer, realScanner)

	handlers := server.NewHandlers(librarySer, seriesSer, bookSer, pageSer, coverSer, thumbnailSer, scanSer)

	if cfg.ScanOnStartup && len(cfg.ScanDirectories) > 0 {
		go func() {
			logger.Info("Starting library scan on startup", "directories", cfg.ScanDirectories)
			if err := scanSer.ScanDirectories(ctx); err != nil {
				logger.Error(err, "Startup scan failed")
			} else {
				logger.Info("Startup scan completed successfully")
			}
		}()
	}

	srv := server.NewServer(ctx, cfg, handlers)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Info("Shutting down server")
		srv.Shutdown()
	}()

	logger.Info("Server starting", "host", cfg.Host)
	if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(err, "Failed to start server")
		os.Exit(1)
	}
	logger.Info("Server stopped")
}
