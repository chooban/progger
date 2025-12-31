package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
)

// DownloadSession holds a browser context and login state for efficient bulk operations.
// Create a session with NewDownloadSession, use it for multiple operations, then call Close.
type DownloadSession struct {
	ctx      context.Context
	bContext playwright.BrowserContext
	details  RebellionDetails
	logger   logr.Logger
}

// NewDownloadSession creates a new download session with an authenticated browser context.
// The caller is responsible for calling Close() when done to clean up resources.
func NewDownloadSession(ctx context.Context, details RebellionDetails) (*DownloadSession, error) {
	logger := logr.FromContextOrDiscard(ctx)

	bContext, err := browser(ctx)
	if err != nil {
		logger.Error(err, "Could not start browser")
		return nil, fmt.Errorf("could not start browser: %w", err)
	}

	if err = Login(ctx, bContext, details.Username, details.Password); err != nil {
		logger.Error(err, "Failed to login")
		bContext.Close()
		return nil, fmt.Errorf("could not login: %w", err)
	}

	return &DownloadSession{
		ctx:      ctx,
		bContext: bContext,
		details:  details,
		logger:   logger,
	}, nil
}

// Close closes the browser context and cleans up resources.
func (s *DownloadSession) Close() error {
	if s.bContext != nil {
		return s.bContext.Close()
	}
	return nil
}

// ListAvailableIssues lists available digital comics using the session's browser context.
func (s *DownloadSession) ListAvailableIssues(latestOnly bool) ([]DigitalComic, error) {
	progs, err := listProgs(s.ctx, s.bContext, latestOnly)
	if err != nil {
		s.logger.V(1).Error(err, "Could not list progs")
		return []DigitalComic{}, err
	}
	return progs, nil
}

// ListIssuesOnPage lists issues on a specific page using the session's browser context.
func (s *DownloadSession) ListIssuesOnPage(pageNumber int) ([]DigitalComic, error) {
	return listIssuesOnPage(s.ctx, s.bContext, pageNumber)
}

// Download downloads a digital comic to the specified directory using the session's browser context.
func (s *DownloadSession) Download(comic DigitalComic, dir string, filetype FileType) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	if _, err = os.Stat(path.Join(dir, comic.Filename(filetype))); err == nil {
		s.logger.V(1).Info("file already exists", "path", path.Join(dir, comic.Filename(filetype)))
		return path.Join(dir, comic.Filename(filetype)), nil
	}

	downloadedFile, err := downloadComic(s.ctx, s.bContext, comic)
	if err != nil {
		return "", fmt.Errorf("failed to downloadComic file: %w", err)
	}

	destinationFile := path.Join(dir, comic.Filename(filetype))

	r, err := os.Open(downloadedFile)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer r.Close()

	w, err := os.Create(destinationFile)
	if err != nil {
		r.Close()
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return destinationFile, nil
}

func ListIssuesOnPage(ctx context.Context, details RebellionDetails, pageNumber int) (issues []DigitalComic, err error) {
	session, err := NewDownloadSession(ctx, details)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session.ListIssuesOnPage(pageNumber)
}

func ListAvailableIssues(ctx context.Context, details RebellionDetails, latestOnly bool) ([]DigitalComic, error) {
	session, err := NewDownloadSession(ctx, details)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session.ListAvailableIssues(latestOnly)
}

func Download(ctx context.Context, details RebellionDetails, comic DigitalComic, dir string, filetype FileType) (string, error) {
	session, err := NewDownloadSession(ctx, details)
	if err != nil {
		return "", err
	}
	defer session.Close()

	return session.Download(comic, dir, filetype)
}

func WithBrowserContextDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, ContextKeyBrowserContext, dir)
}
