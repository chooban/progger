package services

import (
	"context"
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

func (d *Downloader) DownloadIssue(ctx context.Context, issue download.DigitalComic, targetDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
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

func (d *Downloader) DownloadAllIssues(ctx context.Context, sourceDir, username, password string) error {
	logger := logr.FromContextOrDiscard(ctx)
	ctxt := download.WithBrowserContextDir(ctx, d.browserDir)

	details := download.RebellionDetails{
		Username: username,
		Password: password,
	}

	if list, err := download.ListAvailableIssues(ctxt, details, false); err == nil {
		if len(list) > 0 {
			//logger.Info("Found progs to download", "count", len(list))
			//for i := 0; i < len(list); i++ {
			for i := 0; i < 10; i++ {
				logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
				if fp, err := download.Download(ctxt, details, list[i], sourceDir, download.Pdf); err != nil {
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

// StartFetchIssuesList fetches the list of available issues and returns an observable operation
func (d *Downloader) StartFetchIssuesList(username, password, progSourceDir, megSourceDir string) *DownloadListOperation {
	op := NewDownloadListOperation()

	ctx, cancel := context.WithCancel(context.Background())
	op.cancel = cancel

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		// Check if cancelled
		select {
		case <-ctx.Done():
			_ = op.Error.Set("Operation cancelled")
			return
		default:
		}

		ctxt := download.WithBrowserContextDir(ctx, d.browserDir)
		details := download.RebellionDetails{
			Username: username,
			Password: password,
		}

		list, err := download.ListAvailableIssues(ctxt, details, false)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		downloadableList := make([]api.Downloadable, 0, len(list))
		for _, v := range list {
			p := api.Downloadable{
				Comic:      v,
				Downloaded: false,
			}
			downloadableList = append(downloadableList, p)
		}

		progs := BuildIssueList(downloadableList, progSourceDir, megSourceDir)
		if err := op.AvailableProgs.Set(progs); err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Store the issues
		if d.storage != nil {
			if err := d.storage.SaveIssues(downloadableList); err != nil {
				_ = op.Error.Set("Failed to save issues: " + err.Error())
			}
		}
	}()

	return op
}

// StartDownloadIssues downloads the provided issues and returns an observable operation
func (d *Downloader) StartDownloadIssues(issues []api.Downloadable, username, password, progSourceDir, megSourceDir string) *DownloadOperation {
	op := NewDownloadOperation()

	ctx, cancel := context.WithCancel(context.Background())
	op.cancel = cancel

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		for _, v := range issues {
			// Check if cancelled
			select {
			case <-ctx.Done():
				_ = op.Error.Set("Download cancelled")
				return
			default:
			}

			targetDir := progSourceDir
			if v.Comic.Publication == "Megazine" {
				targetDir = megSourceDir
			}

			if err := d.DownloadIssue(ctx, v.Comic, targetDir, username, password); err != nil {
				_ = op.Error.Set(err.Error())
				return
			}
		}
	}()

	return op
}

// StartDownloadAllIssues downloads all available issues and returns an observable operation
func (d *Downloader) StartDownloadAllIssues(username, password, sourceDir string) *DownloadOperation {
	op := NewDownloadOperation()

	ctx, cancel := context.WithCancel(context.Background())
	op.cancel = cancel

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		if err := d.DownloadAllIssues(ctx, sourceDir, username, password); err != nil {
			_ = op.Error.Set(err.Error())
		}
	}()

	return op
}

func NewDownloader(storageRoot string, storage *Storage) *Downloader {
	browserDir := path.Join(storageRoot, "browser")

	return &Downloader{
		browserDir: browserDir,
		storage:    storage,
	}
}
