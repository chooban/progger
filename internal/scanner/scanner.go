package scanner

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/stringutils"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
}

// ScanDir scans the given directory for PDF files and extracts episode details from each file.
// It returns a slice of RawEpisode structs containing the extracted details.
func ScanDir(appEnv env.AppEnv, dir string) (issues []db.Issue) {
	files := getFiles(appEnv, dir)

	jobs := make(chan string, 10)
	results := make(chan db.Issue, len(files))

	var wg sync.WaitGroup

	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go scanWorker(appEnv, &wg, jobs, results)
	}

	for _, file := range files {
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

	return issues
}

func shouldIncludeIssue(issue db.Issue) bool {
	return issue.IssueNumber != 0
}

// ScanFile scans the given file in the specified directory and extracts episode details.
// It returns a slice of RawEpisode structs containing the extracted details and an error if any occurred during the process.
func ScanFile(appEnv env.AppEnv, fileName string) (db.Issue, error) {
	log := appEnv.Log
	pdfcpuConf := model.NewDefaultConfiguration()
	pdfcpuConf.ValidationMode = model.ValidationNone

	if !strings.HasSuffix(fileName, "pdf") {
		return db.Issue{}, errors.New("only pdf files supported")
	}

	f, err := os.Open(fileName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file")
		os.Exit(1)
	}
	log.Debug().Msg("Reading " + f.Name())
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing file")
		}
	}()

	bookmarks, err := api.Bookmarks(f, pdfcpuConf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read bookmarks")
		return db.Issue{}, errors.New("failed to read bookmarks")
	}

	issue := buildIssue(appEnv, fileName, bookmarks)

	return issue, nil
}

func getProgNumber(inFile string) (int, error) {
	filename := filepath.Base(inFile)
	//regex := pcre.MustCompile(`([^()])(\d{,4})\1`, 0)
	regex := regexp.MustCompile(`(\b[^()])(?P<issue>\d{1,4})(\b[^()])`)

	namedResults := stringutils.FindNamedMatches(regex, filename)
	if len(namedResults) > 0 {
		return strconv.Atoi(stringutils.TrimNonAlphaNumeric(namedResults["issue"]))
	}
	return 0, errors.New("no number found in filename")
}

func getFiles(appEnv env.AppEnv, dir string) (pdfFiles []fs.DirEntry) {
	log := appEnv.Log
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Error().Err(err).Msg("Could not read directory")
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
	log := appEnv.Log
	for {
		j, isChannelOpen := <-jobs
		if !isChannelOpen {
			break
		}
		issue, err := ScanFile(appEnv, j)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Failed to read file: %s", j))
		}
		results <- issue
	}
	log.Debug().Msg("Shutting down worker")
	wg.Done()
}
