package scan

import (
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/env"
	"github.com/chooban/progger/scan/internal"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/chooban/progger/scan/types"
	"io/fs"
	"os"
	"strings"
	"sync"
)

// Dir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of Episode structs containing the extracted details.
func Dir(appEnv env.AppEnv, dir string, scanCount int) (issues []types.Issue) {
	files := getFiles(appEnv, dir)

	jobs := make(chan string, 10)
	results := make(chan types.Issue, len(files))

	var wg sync.WaitGroup

	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go scanWorker(appEnv, &wg, jobs, results)
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
func File(appEnv env.AppEnv, fileName string) (types.Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return types.Issue{}, errors.New("only pdf files supported")
	}

	appEnv.Log.Debug().Msg(fmt.Sprintf("Scanning %s", fileName))
	p := pdfium.NewPdfiumReader(appEnv.Log)
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

	issue := internal.BuildIssue(appEnv, fileName, episodeDetails)

	return issue, nil
}

func Credits(appEnv env.AppEnv, fileName string, startingPage int, endingPage int) (types.Credits, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return types.Credits{}, errors.New("only pdf files supported")
	}

	p := pdfium.NewPdfiumReader(appEnv.Log)

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

func scanWorker(appEnv env.AppEnv, wg *sync.WaitGroup, jobs <-chan string, results chan<- types.Issue) {
	for {
		j, isChannelOpen := <-jobs
		if !isChannelOpen {
			break
		}
		issue, err := File(appEnv, j)
		if err != nil {
			appEnv.Log.Error().Err(err).Msg(fmt.Sprintf("Failed to read file: %s", j))
		}
		results <- issue
	}
	appEnv.Log.Debug().Msg("Shutting down worker")
	wg.Done()
}

func shouldIncludeIssue(issue types.Issue) bool {
	return issue.IssueNumber != 0
}
