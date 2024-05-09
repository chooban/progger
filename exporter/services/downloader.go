package services

import (
	"context"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/download"
	downloadApi "github.com/chooban/progger/download/api"
	"github.com/go-logr/logr"
	"os"
	"path"
)

type Downloader struct {
	ctxt           context.Context
	BoundSourceDir binding.String
}

func (d *Downloader) Download(sourceDir, username, password string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	browserDir := path.Join(configDir, "proggerbrowser")

	logger := logr.FromContextOrDiscard(d.ctxt)
	ctx := download.WithLoginDetails(d.ctxt, username, password)
	ctx = download.WithBrowserContextDir(ctx, browserDir)

	list, err := download.ListAvailableProgs(ctx)
	if err != nil {
		return err
	}
	downloadCount := 5
	for i := 0; i < downloadCount && i < len(list); i++ {
		logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
		if fp, err := download.Download(ctx, list[i], sourceDir, downloadApi.Pdf); err != nil {
			logger.Error(err, "could not download file")
		} else {
			logger.Info("Downloaded a file", "file", fp)
		}
	}

	return nil
}

func NewDownloader(ctx context.Context, srcDir binding.String) *Downloader {
	return &Downloader{
		ctxt:           ctx,
		BoundSourceDir: srcDir,
	}
}
