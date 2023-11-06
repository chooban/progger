package pdf

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/rs/zerolog"
	"os"
)

type Reader interface {
	ReadBookmarks(filename string) ([]internal.Bookmark, error)
	Build(episodes []db.Episode)
}

func NewPdfiumReader(log *zerolog.Logger) *PdfiumReader {
	return &PdfiumReader{
		Log:      log,
		Instance: Instance,
	}
}

type PdfiumReader struct {
	Log      *zerolog.Logger
	Instance pdfium.Pdfium
}

func (p *PdfiumReader) ReadBookmarks(filename string) ([]internal.Bookmark, error) {
	// Open the PDF using PDFium (and claim a worker)
	contents, err := os.ReadFile(filename)
	doc, err := p.Instance.OpenDocument(&requests.OpenDocument{
		File: &contents,
	})
	if err != nil {
		p.Log.Err(err).Msg("Could not open file with pdfium")
		return nil, errors.New("failed to read bookmarks")
	}

	defer p.Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	pdfiumBookmarks, err := p.Instance.GetBookmarks(&requests.GetBookmarks{
		Document: doc.Document,
	})
	pageCount, err := p.Instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	bookmarks := make([]internal.Bookmark, len(pdfiumBookmarks.Bookmarks))

	for i, v := range pdfiumBookmarks.Bookmarks {
		b := internal.Bookmark{
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
	return bookmarks, nil

}

// BuildPdf will export a PDF of the provided series and optional
// episodes.
// The parameters of seriesTitle and episodeTitle should be used to
// query the database via appEnv.Db to retrieve all applicable episodes,
// ordered by issue number.
func (p *PdfiumReader) Build(episodes []db.Episode) {

	destination, err := p.Instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
	if err != nil {
		p.Log.Err(err).Msg("Could not create new document")
	}

	pageCount := 0
	for _, v := range episodes {
		filename := fmt.Sprintf("/Users/ross/Documents/2000AD/%s", v.Issue.Filename)
		source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
			Path: &filename,
		})
		if err != nil {
			p.Log.Err(err).Msg(fmt.Sprintf("Could not open PDF: %s", filename))
			continue
		}

		pageRange := fmt.Sprintf("%d-%d", v.PageFrom, v.PageThru)
		_, err = p.Instance.FPDF_ImportPages(&requests.FPDF_ImportPages{
			Source:      source.Document,
			Destination: destination.Document,
			PageRange:   &pageRange,
			Index:       pageCount,
		})
		if err != nil {
			p.Log.Err(err).Msg("Could not import pages")
			continue
		}

		pageCount = (v.PageThru - v.PageFrom) + 1
	}

	var output = "myfile.pdf"
	if saveAsCopy, err := p.Instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    1,
		Document: destination.Document,
		FilePath: &output,
	}); err != nil {
		p.Log.Err(err).Msg("Could not save document")
	} else {
		p.Log.Info().Msg(fmt.Sprintf("File saved to %s", *saveAsCopy.FilePath))
	}

}
