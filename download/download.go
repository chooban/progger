package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/download/api"
	"github.com/chooban/progger/download/internal"
	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
	"io"
	"os"
	"path"
	"slices"
)

func ListAvailableProgs(ctx context.Context) ([]api.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)

	bContext, err := browser(ctx)
	defer func(bContext playwright.BrowserContext) {
		err := bContext.Close()
		if err != nil {
			logger.Error(err, "failed to close browsr")
		}
	}(bContext)

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	}
	u, p := LoginDetails(ctx)

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		return []api.DigitalComic{}, err
	}

	if progs, err := internal.ListProgs(ctx, bContext); err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	} else {
		slices.SortFunc(progs, func(i, j api.DigitalComic) int {
			return j.IssueNumber - i.IssueNumber
		})
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
		return "", fmt.Errorf("file already exists: %s", path.Join(dir, comic.Filename(filetype)))
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
		return "", fmt.Errorf("could not start browser", err)
	}
	u, p := LoginDetails(ctx)

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		return "", fmt.Errorf("could not login", err)
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
