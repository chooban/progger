package services

import (
	"context"
	"github.com/chooban/progger/download"
	downloadApi "github.com/chooban/progger/download/api"
	"github.com/go-logr/logr"
	"os"
	"path"
)

type Downloader struct {
	ctxt context.Context
	//BoundSourceDir binding.String
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
		logger.Error(err, "failed to list available progs")
		return err
	}
	if len(list) > 0 {
		logger.Info("Found progs to download", "count", len(list))
		for i := 0; i < len(list); i++ {
			logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
			if fp, err := download.Download(ctx, list[i], sourceDir, downloadApi.Pdf); err != nil {
				logger.Error(err, "could not download file")
			} else {
				logger.Info("Downloaded a file", "file", fp)
			}
		}
	} else {
		logger.Info("No progs to download")
	}

	return nil
}

func NewDownloader(ctx context.Context) *Downloader {
	return &Downloader{
		ctxt: ctx,
		//BoundSourceDir: srcDir,
	}
}
