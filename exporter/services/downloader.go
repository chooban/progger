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
	browserDir string
}

func (d *Downloader) ProgList(ctx context.Context, username, password string) ([]downloadApi.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithLoginDetails(ctx, username, password)
	ctxt = download.WithBrowserContextDir(ctxt, d.browserDir)

	list, err := download.ListAvailableProgs(ctxt)
	if err != nil {
		logger.Error(err, "failed to list available progs")
		return nil, err
	}

	return list, nil
}

func (d *Downloader) DownloadAllProgs(ctx context.Context, sourceDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithLoginDetails(ctx, username, password)
	ctxt = download.WithBrowserContextDir(ctx, d.browserDir)

	list, err := download.ListAvailableProgs(ctxt)
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

func NewDownloader() *Downloader {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic("could not get user config dir")
	}
	browserDir := path.Join(configDir, "proggerbrowser")

	return &Downloader{
		browserDir: browserDir,
	}
}
