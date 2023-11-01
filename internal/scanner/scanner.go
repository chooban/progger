package scanner

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/stringutils"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rs/zerolog/log"
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

	for _, e := range issues[0].Episodes {
		appEnv.Log.Info().Msg("Attempting to get text")
		getEpisodeText(appEnv, dir+string(os.PathSeparator)+issues[0].Filename, e.PageFrom, e.PageThru)
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

	//bookmarks, err := pdfcpuBookmarks(appEnv, fileName)
	bookmarks, err := tryWithPdfium(appEnv, fileName)
	if err != nil {
		return db.Issue{}, err
	}
	issue := buildIssue(appEnv, fileName, bookmarks)

	return issue, nil
}

func pdfcpuBookmarks(appEnv env.AppEnv, filename string) ([]Bookmark, error) {
	pdfcpuConf := model.NewDefaultConfiguration()
	pdfcpuConf.ValidationMode = model.ValidationNone

	f, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file")
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing file")
		}
	}()

	pdfcpuBookmarks, err := api.Bookmarks(f, pdfcpuConf)
	pageCount, err := api.PageCount(f, pdfcpuConf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read bookmarks")
		return nil, errors.New("failed to read bookmarks")
	}
	bookmarks := make([]Bookmark, len(pdfcpuBookmarks))
	for i, v := range pdfcpuBookmarks {
		b := Bookmark{
			Title:    v.Title,
			PageFrom: v.PageFrom,
			PageThru: v.PageThru,
		}
		if b.PageThru == 0 {
			b.PageThru = pageCount
		}
		bookmarks[i] = b
	}
	return bookmarks, nil
}

func tryWithPdfium(appEnv env.AppEnv, filename string) ([]Bookmark, error) {
	// Open the PDF using PDFium (and claim a worker)
	contents, err := os.ReadFile(filename)
	doc, err := appEnv.Pdfium.OpenDocument(&requests.OpenDocument{
		File: &contents,
	})
	if err != nil {
		appEnv.Log.Err(err).Msg("Could not open file with pdfium")
		return nil, errors.New("failed to read bookmarks")
	}

	// Always close the document, this will release its resources.
	defer appEnv.Pdfium.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	pdfiumBookmarks, err := appEnv.Pdfium.GetBookmarks(&requests.GetBookmarks{
		Document: doc.Document,
	})
	pageCount, err := appEnv.Pdfium.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	bookmarks := make([]Bookmark, len(pdfiumBookmarks.Bookmarks))

	for i, v := range pdfiumBookmarks.Bookmarks {
		b := Bookmark{
			Title:    v.Title,
			PageFrom: v.DestInfo.PageIndex + 1, // It's zero indexed
		}

		if i < len(bookmarks)-1 {
			b.PageThru = pdfiumBookmarks.Bookmarks[i+1].DestInfo.PageIndex
		} else {
			b.PageThru = pageCount.PageCount
		}
		bookmarks[i] = b
	}
	//appEnv.Log.Info().Msg(fmt.Sprintf("%+v", bookmarks))
	return bookmarks, nil
}

func getEpisodeText(appEnv env.AppEnv, filename string, pageFrom int, pageThru int) {
	appEnv.Log.Info().Msg(fmt.Sprintf("Attempting to open %s", filename))
	contents, err := os.ReadFile(filename)
	doc, err := appEnv.Pdfium.OpenDocument(&requests.OpenDocument{
		File: &contents,
	})
	if err != nil {
		appEnv.Log.Err(err).Msg("Could not open file with pdfium")
		return
	}
	// Always close the document, this will release its resources.
	defer appEnv.Pdfium.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	text, _ := appEnv.Pdfium.GetPageText(&requests.GetPageText{
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc.Document,
				Index:    pageFrom - 1,
			},
		},
	})
	appEnv.Log.Info().Msg(text.Text)
	//text, _ := appEnv.Pdfium.GetPageTextStructured(&requests.GetPageTextStructured{
	//	Page: requests.Page{
	//		ByIndex: &requests.PageByIndex{
	//			Document: doc.Document,
	//			Index:    pageFrom - 1,
	//		},
	//	},
	//})

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
