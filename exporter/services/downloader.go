package services

import (
	"context"
	"os"

	"github.com/chooban/progger/download"
	"github.com/chooban/progger/exporter/api"
	"github.com/go-logr/logr"
	"path"
)

type Downloader struct {
	browserDir string
	storage    *Storage
}

func (d *Downloader) GetIssuesList(ctx context.Context, username, password string) ([]download.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithBrowserContextDir(ctx, d.browserDir)

	details := download.RebellionDetails{
		Username: username,
		Password: password,
	}

	if list, err := download.ListAvailableIssues(ctxt, details, false); err == nil {
		return list, nil
	} else {
		logger.Error(err, "failed to list available issues")
		return nil, err
	}
}

// FetchIssuesList fetches the list of available issues and returns them as Downloadables
func (d *Downloader) FetchIssuesList(ctx context.Context, username, password string) ([]api.Downloadable, error) {
	list, err := d.GetIssuesList(ctx, username, password)
	if err != nil {
		return nil, err
	}

	downloadableList := make([]api.Downloadable, 0, len(list))
	for _, v := range list {
		p := api.Downloadable{
			Comic:      v,
			Downloaded: false,
		}
		downloadableList = append(downloadableList, p)
	}

	// Store the issues if storage is available
	if d.storage != nil {
		if err := d.storage.SaveIssues(downloadableList); err != nil {
			// Log but don't fail - storage is non-critical
			logger := logr.FromContextOrDiscard(ctx)
			logger.Error(err, "failed to save issues to storage")
		}
	}

	return downloadableList, nil
}

func (d *Downloader) DownloadIssue(ctx context.Context, issue download.DigitalComic, targetDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)

	destPath := path.Join(targetDir, issue.Filename(download.Pdf))
	if _, err := os.Stat(destPath); err == nil {
		logger.V(1).Info("file already exists, skipping", "path", destPath)
		return nil
	}

	ctxt := download.WithBrowserContextDir(ctx, d.browserDir)

	details := download.RebellionDetails{
		Username: username,
		Password: password,
	}

	if fp, err := download.Download(ctxt, details, issue, targetDir, download.Pdf); err != nil {
		logger.Error(err, "could not download file")
		return err
	} else {
		logger.Info("Downloaded a file", "file", fp)
	}
	return nil
}

// DownloadIssues downloads multiple issues, respecting context cancellation
func (d *Downloader) DownloadIssues(ctx context.Context, issues []api.Downloadable, progSourceDir, megSourceDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithBrowserContextDir(ctx, d.browserDir)

	details := download.RebellionDetails{
		Username: username,
		Password: password,
	}

	session, err := download.NewSession(ctxt, details)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, v := range issues {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		targetDir := progSourceDir
		if v.Comic.Publication == "Megazine" {
			targetDir = megSourceDir
		}

		if fp, err := session.Download(v.Comic, targetDir, download.Pdf); err != nil {
			logger.Error(err, "could not download file")
			return err
		} else {
			logger.Info("Downloaded a file", "file", fp)
		}
	}
	return nil
}

func (d *Downloader) DownloadAllIssues(ctx context.Context, sourceDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithBrowserContextDir(ctx, d.browserDir)

	details := download.RebellionDetails{
		Username: username,
		Password: password,
	}

	session, err := download.NewSession(ctxt, details)
	if err != nil {
		return err
	}
	defer session.Close()

	list, err := session.ListAvailableIssues(false)
	if err != nil {
		logger.Error(err, "failed to list available issues")
		return err
	}

	if len(list) == 0 {
		logger.Info("No issues to download")
		return nil
	}

	for i := 0; i < 10; i++ {
		logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
		if fp, err := session.Download(list[i], sourceDir, download.Pdf); err != nil {
			logger.Error(err, "could not download file")
		} else {
			logger.Info("Downloaded a file", "file", fp)
		}
	}

	return nil
}

func NewDownloader(storageRoot string, storage *Storage) *Downloader {
	browserDir := path.Join(storageRoot, "browser")

	return &Downloader{
		browserDir: browserDir,
		storage:    storage,
	}
}
