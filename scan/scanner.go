package scan

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal"
	"github.com/go-logr/logr"
)

// Scanner encapsulates scanning configuration and operations
type Scanner struct {
	knownSeries []string
	skipTitles  []string
}

// NewScanner creates a new Scanner with the given configuration
func NewScanner(knownSeries, skipTitles []string) *Scanner {
	return &Scanner{
		knownSeries: knownSeries,
		skipTitles:  skipTitles,
	}
}

// Dir scans the given directory for PDF files and extracts episode details from each file.
func (s *Scanner) Dir(ctx context.Context, dir string, scanCount int) ([]api.Issue, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("Scanning directory", "dir", dir)

	files, err := getFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("getting files: %w", err)
	}

	if len(files) == 0 {
		return []api.Issue{}, nil
	}
	logger.Info("Found files to scan", "num_files", len(files))

	jobs := make(chan string, 10)
	results := make(chan api.Issue, len(files))

	var wg sync.WaitGroup

	workerCount := runtime.NumCPU()
	logger.V(1).Info("Creating workers", "num_workers", workerCount)

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go s.scanWorker(ctx, &wg, jobs, results)
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

	issues := make([]api.Issue, 0, len(files))
	for v := range results {
		if v.IssueNumber != 0 {
			issues = append(issues, v)
		}
	}

	// Sanitise the results to correct titles
	Sanitise(ctx, &issues, s.knownSeries)

	return issues, nil
}

// File scans a single PDF file and extracts episode details.
func (s *Scanner) File(ctx context.Context, fileName string) (api.Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return api.Issue{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info(fmt.Sprintf("Scanning %s", fileName))
	p := internal.NewPdfiumReader(logger)
	episodeDetails, err := p.Bookmarks(fileName)
	if err != nil {
		return api.Issue{}, err
	}

	for i := range episodeDetails {
		details := episodeDetails[i]
		if credits, err := p.Credits(fileName, details.Bookmark.PageFrom, details.Bookmark.PageThru); err == nil {
			episodeDetails[i].Credits = credits
		} else {
			logger.V(1).Info("Failed to extract credits", "file", fileName)
		}
	}

	issue := internal.BuildIssue(logger, fileName, episodeDetails, s.knownSeries, s.skipTitles)

	return issue, nil
}

func (s *Scanner) scanWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- api.Issue) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info("Creating worker")
	defer wg.Done()

	for j := range jobs {
		issue, err := s.File(ctx, j)
		if err != nil {
			logger.Error(err, "Failed to read file", "file", j)
		}
		results <- issue
	}
	logger.V(1).Info("Shutting down worker")
}

// Dir scans the given directory for PDF files and extracts episode details from each file.
// Deprecated: Use NewScanner and Scanner.Dir instead for better control.
func Dir(ctx context.Context, dir string, scanCount int, knownSeries []string, skipTitles []string) ([]api.Issue, error) {
	s := NewScanner(knownSeries, skipTitles)
	return s.Dir(ctx, dir, scanCount)
}

// File scans the given file in the specified directory and extracts episode details.
// Deprecated: Use NewScanner and Scanner.File instead for better control.
func File(ctx context.Context, fileName string, knownSeries []string, skipTitles []string) (api.Issue, error) {
	s := NewScanner(knownSeries, skipTitles)
	return s.File(ctx, fileName)
}

func ReadCredits(ctx context.Context, fileName string, startingPage int, endingPage int) (api.Credits, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return api.Credits{}, errors.New("only pdf files supported")
	}
	logger := logr.FromContextOrDiscard(ctx)

	p := internal.NewPdfiumReader(logger)

	credits, err := p.Credits(fileName, startingPage, endingPage)

	if err != nil {
		return api.Credits{}, err
	}
	return internal.ExtractCreatorsFromCredits(credits), nil
}

func getFiles(dir string) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pdfFiles := make([]fs.DirEntry, 0, 100)
	for _, f := range files {
		if f.Type().IsRegular() && strings.HasSuffix(strings.ToLower(f.Name()), "pdf") {
			pdfFiles = append(pdfFiles, f)
		}
	}

	return pdfFiles, nil
}
