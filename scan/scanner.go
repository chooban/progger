package scan

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/env"
	"github.com/chooban/progger/scan/internal"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/chooban/progger/scan/types"
	"github.com/go-logr/logr"
	"io/fs"
	"os"
	"strings"
	"sync"
)

// Dir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of Episode structs containing the extracted details.
func Dir(ctx context.Context, dir string, scanCount int) (issues []types.Issue) {
	appEnv := fromContextOrDefaults(ctx)
	files := getFiles(appEnv, dir)

	jobs := make(chan string, 10)
	results := make(chan types.Issue, len(files))

	var wg sync.WaitGroup

	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go scanWorker(ctx, &wg, jobs, results)
	}

	for i, file := range files {
		jobs <- dir + string(os.PathSeparator) + file.Name()
		if scanCount > 0 && i > scanCount {
			break
		}
	}

	close(jobs)
	wg.Wait()
	close(results)

	for v := range results {
		if shouldIncludeIssue(v) {
			issues = append(issues, v)
		}
	}

	return issues
}

// File scans the given file in the specified directory and extracts episode details.
// It returns a slice of Episode structs containing the extracted details and an error if any occurred during the process.
func File(ctx context.Context, fileName string) (types.Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return types.Issue{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)
	appEnv := fromContextOrDefaults(ctx)

	logger.Info(fmt.Sprintf("Scanning %s", fileName))
	p := pdfium.NewPdfiumReader(logger)
	episodeDetails, err := p.Bookmarks(fileName)
	if err != nil {
		return types.Issue{}, err
	}

	for i, _ := range episodeDetails {
		details := episodeDetails[i]
		credits, err := p.Credits(fileName, details.Bookmark.PageFrom, details.Bookmark.PageThru)
		if err != nil {
			continue
		}
		episodeDetails[i].Credits = credits
	}

	issue := internal.BuildIssue(logger, fileName, episodeDetails, appEnv.Known.SeriesTitles, appEnv.Skip.SeriesTitles)

	return issue, nil
}

func ReadCredits(ctx context.Context, fileName string, startingPage int, endingPage int) (types.Credits, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return types.Credits{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)

	p := pdfium.NewPdfiumReader(logger)

	credits, err := p.Credits(fileName, startingPage, endingPage)

	if err != nil {
		return types.Credits{}, err
	}
	return internal.ExtractCreatorsFromCredits(credits), nil
}

func getFiles(appEnv env.AppEnv, dir string) (pdfFiles []fs.DirEntry) {
	files, err := os.ReadDir(dir)
	if err != nil {
		appEnv.Log.Error().Err(err).Msg("Could not read directory")
	}

	pdfFiles = make([]fs.DirEntry, 0, 100)
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(strings.ToLower(f.Name()), "pdf") {
			pdfFiles = append(pdfFiles, f)
		}
	}

	return
}

func scanWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- types.Issue) {
	logger := logr.FromContextOrDiscard(ctx)
	for {
		j, isChannelOpen := <-jobs
		if !isChannelOpen {
			break
		}
		issue, err := File(ctx, j)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to read file: %s", j))
		}
		results <- issue
	}
	logger.Info("Shutting down worker")
	wg.Done()
}

func shouldIncludeIssue(issue types.Issue) bool {
	return issue.IssueNumber != 0
}
