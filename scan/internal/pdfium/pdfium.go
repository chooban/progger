package pdfium

import (
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/internal/pdf"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/rs/zerolog"
	"math"
	"os"
	"slices"
	"strings"
)

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

func (p *PdfiumReader) Bookmarks(filename string) ([]pdf.EpisodeDetails, error) {
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
	bookmarks := make([]pdf.Bookmark, len(pdfiumBookmarks.Bookmarks))

	for i, v := range pdfiumBookmarks.Bookmarks {
		b := pdf.Bookmark{
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
	details := make([]pdf.EpisodeDetails, len(bookmarks))

	for i, v := range bookmarks {
		details[i] = pdf.EpisodeDetails{
			Bookmark: v,
		}
	}
	return details, nil
}

// Build will export a PDF of the provided episodes.
//func (p *PdfiumReader) Build(episodes []types.Episode) {
//
//	destination, err := p.Instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
//	if err != nil {
//		p.Log.Err(err).Msg("Could not create new document")
//	}
//
//	pageCount := 0
//	for _, v := range episodes {
//		filename := fmt.Sprintf("/Users/ross/Documents/2000AD/%s", v.Issue.Filename)
//		source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
//			Path: &filename,
//		})
//		if err != nil {
//			p.Log.Err(err).Msg(fmt.Sprintf("Could not open PDF: %s", filename))
//			continue
//		}
//
//		pageRange := fmt.Sprintf("%d-%d", v.PageFrom, v.PageThru)
//		_, err = p.Instance.FPDF_ImportPages(&requests.FPDF_ImportPages{
//			Source:      source.Document,
//			Destination: destination.Document,
//			PageRange:   &pageRange,
//			Index:       pageCount,
//		})
//		if err != nil {
//			p.Log.Err(err).Msg("Could not import pages")
//			continue
//		}
//
//		pageCount = (v.PageThru - v.PageFrom) + 1
//	}
//
//	var output = "myfile.pdf"
//	if saveAsCopy, err := p.Instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
//		Flags:    1,
//		Document: destination.Document,
//		FilePath: &output,
//	}); err != nil {
//		p.Log.Err(err).Msg("Could not save document")
//	} else {
//		p.Log.Info().Msg(fmt.Sprintf("File saved to %s", *saveAsCopy.FilePath))
//	}
//}

func (p *PdfiumReader) Credits(filename string, startPage int, endPage int) (credits string, err error) {
	source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: &filename,
	})

	defer func() {
		p.Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: source.Document})
	}()

	if err != nil {
		p.Log.Err(err).Msg("Could not open file")
		return "", err
	}

	p.Log.Debug().Msg(fmt.Sprintf("Reading %s", filename))
	var creditTypes = []string{"script", "art", "colours", "letters"}
	var textPage *responses.FPDFText_LoadPage
	var scriptRect *responses.FPDFText_GetRect

	for pageIndex := startPage; pageIndex <= endPage; pageIndex++ {
		p.Log.Debug().Msg(fmt.Sprintf("Scanning page %d of %s", pageIndex, filename))
		if pdfPage, err := p.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: source.Document,
			Index:    pageIndex - 1,
		}); err != nil {
			p.Log.Err(err).Msg(fmt.Sprintf("Failed to load page %d", pageIndex))
			return "", errors.New("failed to load page")
		} else {
			if textPage, scriptRect = p.findScriptRect(pdfPage.Page); scriptRect != nil {
				p.Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{
					Page: pdfPage.Page,
				})
				break
			}
		}
	}
	if scriptRect == nil {
		return "", errors.New("no script found in range")
	}
	var (
		left       = scriptRect.Left - ((scriptRect.Right - scriptRect.Left) * 1.1)
		right      = scriptRect.Right + ((scriptRect.Right - scriptRect.Left) * 1.1)
		top        = scriptRect.Top + 20
		bottom     = scriptRect.Bottom - 20
		rawCredits = "script"
	)

	for bottom >= 0 {
		creditsText, _ := p.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
			TextPage: textPage.TextPage,
			Left:     left,
			Right:    right,
			Bottom:   bottom,
			Top:      top,
		})
		if creditsText.Text != rawCredits {
			rawCredits = creditsText.Text
			bottom -= 20
			continue
		}
		break
	}

	tokenized := strings.Fields(strings.ToLower(strings.ReplaceAll(rawCredits, "\r\n", " ")))
	earliestIdx := math.MaxInt16
	latestIdx := math.MinInt16
	for _, v := range creditTypes {
		cIdx := slices.Index(tokenized, v)
		if cIdx >= 0 {
			earliestIdx = min(cIdx, earliestIdx)
			latestIdx = max(cIdx, latestIdx)
		}
	}
	tmpCredits := tokenized[earliestIdx:min(latestIdx+4, len(tokenized))]

	credits = strings.Join(tmpCredits, " ")

	return credits, nil
}

func (p *PdfiumReader) findScriptRect(pageRef references.FPDF_PAGE) (*responses.FPDFText_LoadPage, *responses.FPDFText_GetRect) {
	var (
		textPage   *responses.FPDFText_LoadPage
		scriptRect *responses.FPDFText_GetRect
	)
	textPage, _ = p.Instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{
		Page: requests.Page{
			ByReference: &pageRef,
			ByIndex:     nil,
		},
	})
	rects, _ := p.Instance.FPDFText_CountRects(&requests.FPDFText_CountRects{
		TextPage:   textPage.TextPage,
		StartIndex: 0,
		Count:      -1,
	})
	for textRectIndex := 0; textRectIndex < rects.Count; textRectIndex++ {
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
			p.Log.Debug().Msg(fmt.Sprintf("Found script at %+v", rect))
			scriptRect = rect
		}
	}
	return textPage, scriptRect
}
