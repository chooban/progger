package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/download/api"
	"github.com/chooban/progger/download/internal"
	"github.com/go-logr/logr"
	"os"
	"path"
)

func ListAvailableProgs(ctx context.Context) ([]api.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)

	bContext, err := browser()
	defer bContext.Close()

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
		return "", fmt.Errorf("file already exists")
	}

	bContext, err := browser()
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

	loc, err := internal.Download(ctx, bContext, comic)
	if err != nil {
		return "", fmt.Errorf("failed to download file %g", err)
	}

	dest := path.Join(dir, comic.Filename(filetype))
	err = os.Rename(loc, dest)
	if err != nil {
		return "", fmt.Errorf("failed to move downloaded file %g", err)
	}

	return dest, err
}
