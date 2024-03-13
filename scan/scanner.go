package scan

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/go-logr/logr"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Dir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of Episode structs containing the extracted details.
func Dir(ctx context.Context, dir string, scanCount int) (issues []api.Issue) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("Scanning directory", "dir", dir)

	files, _ := getFiles(dir)

	if len(files) == 0 {
		return
	}
	logger.Info("Found files to scan", "num_files", len(files))

	jobs := make(chan string, 10)
	results := make(chan api.Issue, len(files))

	var wg sync.WaitGroup

	workerCount := runtime.NumCPU()
	logger.V(1).Info("Creating workers", "num_workers", workerCount)

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go scanWorker(ctx, &wg, jobs, results)
	}

	for i, file := range files {
		if scanCount > 0 && i >= scanCount {
			break
		}
		logger.V(1).Info("Adding file to jobs", "file_name", file.Name())
		jobs <- dir + string(os.PathSeparator) + file.Name()
	}

	close(jobs)
	wg.Wait()
	close(results)

	for v := range results {
		if shouldIncludeIssue(v) {
			issues = append(issues, v)
		}
	}

	// Sanitise the results to correct titles
	Sanitise(ctx, &issues)

	return issues
}

// File scans the given file in the specified directory and extracts episode details.
// It returns a slice of Episode structs containing the extracted details and an error if any occurred during the process.
func File(ctx context.Context, fileName string) (api.Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return api.Issue{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)
	appEnv := fromContextOrDefaults(ctx)

	logger.Info(fmt.Sprintf("Scanning %s", fileName))
	p := pdfium.NewPdfiumReader(logger)
	episodeDetails, err := p.Bookmarks(fileName)
	if err != nil {
		return api.Issue{}, err
	}

	for i, _ := range episodeDetails {
		details := episodeDetails[i]
		credits, err := p.Credits(fileName, details.Bookmark.PageFrom, details.Bookmark.PageThru)
		if err != nil {
			continue
		}
		episodeDetails[i].Credits = credits
	}

	issue := buildIssue(logger, fileName, episodeDetails, appEnv.Known, appEnv.Skip)

	return issue, nil
}

func ReadCredits(ctx context.Context, fileName string, startingPage int, endingPage int) (api.Credits, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return api.Credits{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)

	p := pdfium.NewPdfiumReader(logger)

	credits, err := p.Credits(fileName, startingPage, endingPage)

	if err != nil {
		return api.Credits{}, err
	}
	return extractCreatorsFromCredits(credits), nil
}

func getFiles(dir string) (pdfFiles []fs.DirEntry, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	pdfFiles = make([]fs.DirEntry, 0, 100)
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(strings.ToLower(f.Name()), "pdf") {
			pdfFiles = append(pdfFiles, f)
		}
	}

	return
}

func scanWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- api.Issue) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info("Creating worker")
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
	logger.V(1).Info("Shutting down worker")
	wg.Done()
}

func shouldIncludeIssue(issue api.Issue) bool {
	return issue.IssueNumber != 0
}
