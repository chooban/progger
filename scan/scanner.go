package scan

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/go-logr/logr"
	"io/fs"
	"os"
	"strings"
	"sync"
)

// Dir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of Episode structs containing the extracted details.
func Dir(ctx context.Context, dir string, scanCount int) (issues []Issue) {
	files, _ := getFiles(dir)

	jobs := make(chan string, 10)
	results := make(chan Issue, len(files))

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
func File(ctx context.Context, fileName string) (Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return Issue{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)
	appEnv := fromContextOrDefaults(ctx)

	logger.Info(fmt.Sprintf("Scanning %s", fileName))
	p := pdfium.NewPdfiumReader(logger)
	episodeDetails, err := p.Bookmarks(fileName)
	if err != nil {
		return Issue{}, err
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

func ReadCredits(ctx context.Context, fileName string, startingPage int, endingPage int) (Credits, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return Credits{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)

	p := pdfium.NewPdfiumReader(logger)

	credits, err := p.Credits(fileName, startingPage, endingPage)

	if err != nil {
		return Credits{}, err
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

func scanWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- Issue) {
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

func shouldIncludeIssue(issue Issue) bool {
	return issue.IssueNumber != 0
}
