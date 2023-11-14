package scanner

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/stringutils"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// A RawEpisode represents information extracted from a PDF bookmark. It is not a database record
type RawEpisode struct {
	Series    string
	Title     string
	Part      int
	FirstPage int
	LastPage  int
	Script    []string
	Art       []string
	Colours   []string
	Letters   []string
}

// ScanDir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of RawEpisode structs containing the extracted details.
func ScanDir(appEnv env.AppEnv, dir string, scanCount int) (issues []db.Issue) {
	files := getFiles(appEnv, dir)

	jobs := make(chan string, 10)
	results := make(chan db.Issue, len(files))

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

func shouldIncludeIssue(issue db.Issue) bool {
	return issue.IssueNumber != 0
}

// ScanFile scans the given file in the specified directory and extracts episode details.
// It returns a slice of RawEpisode structs containing the extracted details and an error if any occurred during the process.
func ScanFile(appEnv env.AppEnv, fileName string) (db.Issue, error) {
	if !strings.HasSuffix(fileName, "pdf") {
		return db.Issue{}, errors.New("only pdf files supported")
	}

	episodeDetails, err := appEnv.Pdf.Bookmarks(fileName)
	if err != nil {
		return db.Issue{}, err
	}

	issue := buildIssue(appEnv, fileName, episodeDetails)

	return issue, nil
}

func getProgNumber(inFile string) (int, error) {
	filename := filepath.Base(inFile)
	regex := regexp.MustCompile(`(\b[^()])(?P<issue>\d{1,4})(\b[^()])`)

	namedResults := stringutils.FindNamedMatches(regex, filename)
	if len(namedResults) > 0 {
		return strconv.Atoi(stringutils.TrimNonAlphaNumeric(namedResults["issue"]))
	}
	return 0, errors.New("no number found in filename")
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

func scanWorker(appEnv env.AppEnv, wg *sync.WaitGroup, jobs <-chan string, results chan<- db.Issue) {
	for {
		j, isChannelOpen := <-jobs
		if !isChannelOpen {
			break
		}
		issue, err := ScanFile(appEnv, j)
		if err != nil {
			appEnv.Log.Error().Err(err).Msg(fmt.Sprintf("Failed to read file: %s", j))
		}
		results <- issue
	}
	appEnv.Log.Debug().Msg("Shutting down worker")
	wg.Done()
}
