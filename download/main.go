package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path"
)

func ListIssuesOnPage(ctx context.Context, details RebellionDetails, pageNumber int) (issues []DigitalComic, err error) {
	logger := logr.FromContextOrDiscard(ctx)
	bContext, err := browser(ctx)

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []DigitalComic{}, err
	}
	if err = Login(ctx, bContext, details.Username, details.Password); err != nil {
		logger.Error(err, "Failed to login")
		return []DigitalComic{}, err
	}

	issues, err = listIssuesOnPage(ctx, bContext, pageNumber)

	return issues, nil
}

func ListAvailableIssues(ctx context.Context, details RebellionDetails, latestOnly bool) ([]DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)
	bContext, err := browser(ctx)

	if err != nil {
		logger.V(1).Error(err, "Could not start browser")
		return []DigitalComic{}, err
	}

	if err = Login(ctx, bContext, details.Username, details.Password); err != nil {
		logger.V(1).Error(err, "Failed to login")
		return []DigitalComic{}, err
	}

	if progs, err := listProgs(ctx, bContext, latestOnly); err != nil {
		logger.V(1).Error(err, "Could not list progs")
		return []DigitalComic{}, err
	} else {
		return progs, nil
	}
}

func Download(ctx context.Context, details RebellionDetails, comic DigitalComic, dir string, filetype FileType) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("directory does not exist: %g", err)
	}
	if !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	if _, err = os.Stat(path.Join(dir, comic.Filename(filetype))); err == nil {
		logger.V(1).Info("file already exists", "path", path.Join(dir, comic.Filename(filetype)))
		return path.Join(dir, comic.Filename(filetype)), nil
	}

	bContext, err := browser(ctx)
	defer func() {
		err := bContext.Close()
		if err != nil {
			logger.Error(err, "failed to close browser")
		}
	}()

	if err != nil {
		logger.Error(err, "Could not start browser")
		return "", fmt.Errorf("could not start browser: %w", err)
	}
	if err = Login(ctx, bContext, details.Username, details.Password); err != nil {
		return "", fmt.Errorf("could not login: %w", err)
	}

	downloadedFile, err := downloadComic(ctx, bContext, comic)
	if err != nil {
		return "", fmt.Errorf("failed to downloadComic file %g", err)
	}

	destinationFile := path.Join(dir, comic.Filename(filetype))

	r, _ := os.Open(downloadedFile)
	w, _ := os.Create(destinationFile)
	_, err = io.Copy(w, r)
	if err != nil {
		return "", fmt.Errorf("failed to move downloaded file %g", err)
	}

	return destinationFile, err
}

func WithBrowserContextDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, ContextKeyBrowserContext, dir)
}
