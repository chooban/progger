package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"github.com/chooban/progger/download"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"os"
	"slices"
	"time"
)

type arrayFlags []string

// String is an implementation of the flag.Value interface
func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

// Set is an implementation of the flag.Value interface
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	ctx, logger := withLogger(context.Background())

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error(err, "Could not determine home dir")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Error(err, "Could not determine config dir")
	}

	details := download.RebellionDetails{
		Username: os.Getenv("REBELLION_USERNAME"),
		Password: os.Getenv("REBELLION_PASSWORD"),
	}

	var printUsage bool
	flag.BoolVar(&printUsage, "help", false, "Print usage")

	var listAvailable bool
	flag.BoolVar(&listAvailable, "list", false, "List available downloads")

	var listLatest bool
	flag.BoolVar(&listLatest, "latest", false, "List latest downloads")

	var downloadFiles bool
	flag.BoolVar(&downloadFiles, "download", false, "Download files")

	var listPage = 0
	flag.IntVar(&listPage, "list-page", listPage, "Specific page to scan")

	var browserDir string
	flag.StringVar(&browserDir, "browser-dir", configDir, "Directory for browser cache")

	var downloadDir string
	flag.StringVar(&downloadDir, "download-dir", homeDir, "Directory for downloads")

	var downloadCount = 0
	flag.IntVar(&downloadCount, "download-count", downloadCount, "Number of progs to download")

	var publicationFilter arrayFlags
	flag.Var(&publicationFilter, "publication-filter", "Publication names to download. Defaults to all")

	flag.Parse()

	if printUsage {
		flag.PrintDefaults()
		return
	}

	if _, err := isWritable(downloadDir); err != nil {
		logger.Error(err, "Specified download dir is not writable")
	}

	ctx = download.WithBrowserContextDir(ctx, browserDir)

	var list []download.DigitalComic
	if listPage > 0 {
		list, err = download.ListIssuesOnPage(ctx, details, listPage)
	} else {
		list, err = download.ListAvailableIssues(ctx, details, listLatest)
	}

	if err != nil {
		logger.Error(err, "Error listing progs")
	}

	if len(list) == 0 {
		logger.Info("No progs found")
		return
	}

	issueCmp := func(i, j download.DigitalComic) int {
		return cmp.Compare(j.IssueNumber, i.IssueNumber)
	}
	slices.SortFunc(list, issueCmp)

	if listAvailable {
		for _, prog := range list {
			logger.Info(fmt.Sprintf("Found %s, %d, %s", prog.Publication, prog.IssueNumber, prog.IssueDate))
		}
	}

	for i := 0; downloadFiles && i < downloadCount && i < len(list); i++ {
		if len(publicationFilter) > 0 {
			if !slices.Contains(publicationFilter, list[i].Publication) {
				logger.Info(fmt.Sprintf("Publication %s not found in filter list", list[i].Publication))
				continue
			}
		}
		logger.Info("Downloading issue", "issue_number", list[i].IssueNumber)
		if filepath, err := download.Download(ctx, details, list[i], downloadDir, download.Pdf); err != nil {
			logger.Error(err, "could not download file")
		} else {
			logger.Info("Downloaded a file", "file", filepath)
		}
	}
}

func withLogger(ctx context.Context) (context.Context, logr.Logger) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zlogger := zerolog.New(writer)
	zlogger = zlogger.With().Caller().Logger()

	var logrLogger = zerologr.New(&zlogger)
	ctx = logr.NewContext(ctx, logrLogger)

	return ctx, logrLogger
}

func isWritable(path string) (bool, error) {
	tmpFile := "tmpfile"

	file, err := os.CreateTemp(path, tmpFile)
	if err != nil {
		return false, err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return true, nil
}
