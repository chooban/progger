package pdf

import (
	"errors"
	"fmt"
	"github.com/chooban/progdl-go/internal"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

type Reader interface {
	Bookmarks(filename string) ([]internal.Bookmark, error)
	Build(episodes []db.Episode)
	Credits(filename string, startPage int, endPage int) (string, error)
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

func (p *PdfiumReader) Credits(filename string, startPage int, endPage int) (contents string, err error) {
	source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: &filename,
	})
	var pdfPage *responses.FPDF_LoadPage
	for pageIndex := startPage; pageIndex <= endPage; pageIndex++ {
		if pdfPage, err = p.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: source.Document,
			Index:    startPage - 1,
		}); err != nil {
			p.Log.Err(err).Msg("Failed to load pageIndex")
			return "", errors.New("failed to load pageIndex")
		}
		textPage, _ := p.Instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{
			Page: requests.Page{
				ByIndex:     nil,
				ByReference: &pdfPage.Page,
			}})

		textCounts, _ := p.Instance.FPDFText_CountRects(&requests.FPDFText_CountRects{
			TextPage:   textPage.TextPage,
			StartIndex: 0,
			Count:      -1,
		})

		for textRectIndex := 0; textRectIndex < textCounts.Count; textRectIndex++ {
			rect, _ := p.Instance.FPDFText_GetRect(&requests.FPDFText_GetRect{
				TextPage: textPage.TextPage,
				Index:    textRectIndex,
			})
			text, _ := p.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
				TextPage: textPage.TextPage,
				Left:     rect.Left,
				Top:      rect.Top,
				Right:    rect.Right,
				Bottom:   rect.Bottom,
			})
			if strings.ToLower(text.Text) == "script" {
				// Try to find a pageIndex object that contains this text
				if pathBB, err := p.findSurroundingPath(&pdfPage.Page, rect); err == nil {
					creditsText, _ := p.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
						TextPage: textPage.TextPage,
						Left:     float64(pathBB.Left),
						Top:      float64(pathBB.Top),
						Right:    float64(pathBB.Right),
						Bottom:   float64(pathBB.Left),
					})
					contents = strings.ReplaceAll(creditsText.Text, "\r\n", " ")
				}
			}
		}
	}
	return contents, nil
}

func (p *PdfiumReader) findSurroundingPath(pageRef *references.FPDF_PAGE, textRect *responses.FPDFText_GetRect) (*responses.FPDFPageObj_GetBounds, error) {
	objCounts, _ := p.Instance.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{Page: requests.Page{
		ByReference: pageRef,
	}})
	for i := 0; i < objCounts.Count; i++ {
		pageObj, _ := p.Instance.FPDFPage_GetObject(&requests.FPDFPage_GetObject{
			Page: requests.Page{
				ByReference: pageRef,
			},
			Index: i,
		})

		pageObjType, _ := p.Instance.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: pageObj.PageObject})
		bb, _ := p.Instance.FPDFPageObj_GetBounds(&requests.FPDFPageObj_GetBounds{PageObject: pageObj.PageObject})
		if pageObjType.Type != 1 && pageObjType.Type != 3 && bbContains(*bb, *textRect) {
			return bb, nil
		}
	}
	return &responses.FPDFPageObj_GetBounds{
		Left:   0,
		Right:  0,
		Top:    0,
		Bottom: 0,
	}, errors.New("could not find surrounding path")
}

// bbContains returns true if the textBb is contained completely within objBb
func bbContains(objBb responses.FPDFPageObj_GetBounds, textBb responses.FPDFText_GetRect) bool {
	return float32(textBb.Left) >= objBb.Left &&
		float32(textBb.Right) <= objBb.Right &&
		float32(textBb.Top) <= objBb.Top &&
		float32(textBb.Bottom) >= objBb.Bottom
}
