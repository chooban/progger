package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/download/api"
	"github.com/chooban/progger/download/internal"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path"
)

func ListIssuesOnPage(ctx context.Context, pageNumber int) (issues []api.DigitalComic, err error) {
	logger := logr.FromContextOrDiscard(ctx)
	bContext, err := browser(ctx)

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	}
	u, p, err := LoginDetails(ctx)

	if err != nil {
		logger.Error(err, "no credentials found")
		return []api.DigitalComic{}, err
	}

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		logger.Error(err, "Failed to login")
		return []api.DigitalComic{}, err
	}

	issues, err = internal.ListIssuesOnPage(ctx, bContext, pageNumber)

	return issues, nil
}

func ListAvailableIssues(ctx context.Context, latestOnly bool) ([]api.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)
	bContext, err := browser(ctx)

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	}
	u, p, err := LoginDetails(ctx)

	if err != nil {
		logger.Error(err, "no credentials found")
		return []api.DigitalComic{}, err
	}

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		logger.Error(err, "Failed to login")
		return []api.DigitalComic{}, err
	}

	if progs, err := internal.ListProgs(ctx, bContext, latestOnly); err != nil {
		logger.Error(err, "Could not list progs")
		return []api.DigitalComic{}, err
	} else {
		//slices.SortFunc(progs, func(i, j api.DigitalComic) int {
		//	return j.IssueNumber - i.IssueNumber
		//})
		return progs, nil
	}
}

func Download(ctx context.Context, comic api.DigitalComic, dir string, filetype api.FileType) (string, error) {
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
	u, p, err := LoginDetails(ctx)
	if err != nil {
		logger.Error(err, "no credentials found")
		return "", errors.New("no credentials found")
	}

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		return "", fmt.Errorf("could not login: %w", err)
	}

	downloadedFile, err := internal.Download(ctx, bContext, comic)
	if err != nil {
		return "", fmt.Errorf("failed to download file %g", err)
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
