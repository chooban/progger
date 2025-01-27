package services

import (
	"context"
	"github.com/chooban/progger/download"
	"github.com/go-logr/logr"
	"path"
)

type Downloader struct {
	browserDir string
}

func (d *Downloader) GetIssuesList(ctx context.Context, username, password string) ([]download.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithLoginDetails(ctx, username, password)
	ctxt = download.WithBrowserContextDir(ctxt, d.browserDir)

	if list, err := download.ListAvailableIssues(ctxt, false); err == nil {
		return list, nil
	} else {
		logger.Error(err, "failed to list available issues")
		return nil, err
	}
}

func (d *Downloader) DownloadIssue(ctx context.Context, issue download.DigitalComic, targetDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithLoginDetails(ctx, username, password)
	ctxt = download.WithBrowserContextDir(ctxt, d.browserDir)

	if fp, err := download.Download(ctxt, issue, targetDir, download.Pdf); err != nil {
		logger.Error(err, "could not download file")
		return err
	} else {
		logger.Info("Downloaded a file", "file", fp)
	}
	return nil
}

func (d *Downloader) DownloadAllIssues(ctx context.Context, sourceDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithLoginDetails(ctx, username, password)
	ctxt = download.WithBrowserContextDir(ctx, d.browserDir)

	if list, err := download.ListAvailableIssues(ctxt, false); err == nil {
		if len(list) > 0 {
			//logger.Info("Found progs to download", "count", len(list))
			//for i := 0; i < len(list); i++ {
			for i := 0; i < 10; i++ {
				logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
				if fp, err := download.Download(ctxt, list[i], sourceDir, download.Pdf); err != nil {
					logger.Error(err, "could not download file")
				} else {
					logger.Info("Downloaded a file", "file", fp)
				}
			}
		} else {
			logger.Info("No issues to download")
		}
	} else {
		logger.Error(err, "failed to list available issues")
		return err
	}

	return nil
}

func NewDownloader(storageRoot string) *Downloader {
	browserDir := path.Join(storageRoot, "browser")

	return &Downloader{
		browserDir: browserDir,
	}
}
