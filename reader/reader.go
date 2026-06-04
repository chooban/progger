package reader

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/chooban/progger/database"
	"github.com/chooban/progger/reader/server"
	"github.com/chooban/progger/scan"
	"github.com/go-logr/logr"
)

type ServerConfig struct {
	DatabasePath    string
	Host            string
	LibraryName     string
	ScanDirectories []string
	Logger          logr.Logger
}

type Server struct {
	addr    string
	ready   chan struct{}
	errCh   chan error
	httpSrv *server.Server
	db      *database.DB
	mu      sync.Mutex
	running bool
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Ready() <-chan struct{} {
	return s.ready
}

func (s *Server) Err() <-chan error {
	return s.errCh
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	s.httpSrv.Shutdown()
	return s.db.Close()
}

func Start(ctx context.Context, cfg ServerConfig) (*Server, error) {
	if cfg.Host == "" {
		cfg.Host = ":8420"
	}
	if cfg.LibraryName == "" {
		cfg.LibraryName = "2000 AD"
	}

	pdfApi.DisableConfigDir()

	logger := cfg.Logger
	ctx = logr.NewContext(ctx, logger)

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	libraryRepo := database.NewLibraryRepo(db)
	seriesRepo := database.NewSeriesRepo(db)
	bookRepo := database.NewBookRepo(db)
	coverRepo := database.NewCoverRepo(db)
	pageService := server.NewPageService(bookRepo, nil, scan.BuildPageAsPDF, scan.BuildPageAsImage)
	thumbnailService := server.NewThumbnailService(pageService)

	knownTitles, _ := database.NewKnownTitlesRepo(db).List(ctx)
	skipTitles, _ := database.NewSkipTitlesRepo(db).List(ctx)
	realScanner := server.NewRealScanner(knownTitles, skipTitles)

	scanCfg := &server.Config{
		LibraryName:     cfg.LibraryName,
		Host:            cfg.Host,
		ScanDirectories: cfg.ScanDirectories,
	}
	scanService := server.NewScanService(db, scanCfg, libraryRepo, seriesRepo, bookRepo, coverRepo, realScanner)

	handlers := server.NewHandlers(libraryRepo, seriesRepo, bookRepo, pageService, coverRepo, thumbnailService, scanService)

	srv := server.NewServer(ctx, scanCfg, handlers)

	addr := "http://localhost" + cfg.Host

	ready := make(chan struct{})
	errCh := make(chan error, 1)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Info("Shutting down server")
		srv.Shutdown()
	}()

	go func() {
		logger.Info("Server starting", "host", cfg.Host)
		close(ready)
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "Failed to start server")
			errCh <- err
		}
		logger.Info("Server stopped")
	}()

	return &Server{
		addr:    addr,
		ready:   ready,
		errCh:   errCh,
		httpSrv: srv,
		db:      db,
		running: true,
	}, nil
}
