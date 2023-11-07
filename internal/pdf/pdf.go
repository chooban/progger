package pdf

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

type Reader interface {
	Bookmarks(filename string) ([]internal.Bookmark, error)
	Build(episodes []db.Episode)
	Credits(filename string, page int) (string, error)
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

func (p *PdfiumReader) Bookmarks(filename string) ([]internal.Bookmark, error) {
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

// Build will export a PDF of the provided series and optional
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

func (p *PdfiumReader) Credits(filename string, pageNumber int) (contents string, err error) {
	source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: &filename,
	})
	pdfPage, err := p.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
		Document: source.Document,
		Index:    pageNumber - 1,
	})
	if err != nil {
		p.Log.Err(err).Msg("Failed to load page")
		panic(1)
	}
	textPage, err := p.Instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{
		Page: requests.Page{
			ByIndex:     nil,
			ByReference: &pdfPage.Page,
		}})

	objCounts, err := p.Instance.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{Page: requests.Page{
		ByReference: &pdfPage.Page,
	}})
	for i := 0; i < objCounts.Count; i++ {
		pageobj, _ := p.Instance.FPDFPage_GetObject(&requests.FPDFPage_GetObject{
			Page: requests.Page{
				ByReference: &pdfPage.Page,
			},
			Index: i,
		})

		pageObjType, _ := p.Instance.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: pageobj.PageObject})
		bb, _ := p.Instance.FPDFPageObj_GetBounds(&requests.FPDFPageObj_GetBounds{PageObject: pageobj.PageObject})

		p.Log.Info().Msg(fmt.Sprintf("Type: %+v. Bounding box: %+v", pageObjType.Type, bb))
	}

	counts, err := p.Instance.FPDFText_CountRects(&requests.FPDFText_CountRects{
		TextPage:   textPage.TextPage,
		StartIndex: 0,
		Count:      -1,
	})

	for i := 0; i < counts.Count; i++ {
		rect, _ := p.Instance.FPDFText_GetRect(&requests.FPDFText_GetRect{
			TextPage: textPage.TextPage,
			Index:    i,
		})
		text, _ := p.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
			TextPage: textPage.TextPage,
			Left:     rect.Left,
			Top:      rect.Top,
			Right:    rect.Right,
			Bottom:   rect.Bottom,
		})
		if strings.ToLower(text.Text) == "script" {
			p.Log.Info().Msg(fmt.Sprintf("Found the script entry. %+v", rect))
			height := rect.Bottom - rect.Top
			width := rect.Right - rect.Left
			creditsBox, _ := p.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
				TextPage: textPage.TextPage,
				Left:     rect.Left - (width / 2),
				Top:      rect.Top,
				Right:    rect.Right + width + width/2,
				Bottom:   rect.Bottom + (18 * height),
			})
			p.Log.Info().Msg(creditsBox.Text)

			contents = strings.ReplaceAll(creditsBox.Text, "\r\n", " ")

			p.Log.Info().Msg(contents)
		}
	}
	return contents, nil
}

// bbContains returns true if the textBb is contained completely within objBb
func bbContains(objBb responses.FPDFPageObj_GetBounds, textBb responses.FPDFText_GetRect) bool {
	return float32(textBb.Left) >= objBb.Left &&
		float32(textBb.Right) <= objBb.Right &&
		float32(textBb.Top) >= objBb.Top &&
		float32(textBb.Bottom) <= objBb.Bottom
}
