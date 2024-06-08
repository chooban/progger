package pdfium

import (
	"errors"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal/pdf"
	"github.com/go-logr/logr"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"math"
	"os"
	"slices"
	"strings"
)

func NewPdfiumReader(log logr.Logger) *Reader {
	return &Reader{
		Log:      log,
		Instance: Instance,
	}
}

type Reader struct {
	Log      logr.Logger
	Instance pdfium.Pdfium
}

func (p *Reader) Bookmarks(filename string) ([]pdf.EpisodeDetails, error) {
	contents, err := os.ReadFile(filename)
	doc, err := p.Instance.OpenDocument(&requests.OpenDocument{
		File: &contents,
	})
	if err != nil {
		p.Log.Error(err, "Could not open file with pdfium")
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

func (p *Reader) Build(episodes []api.ExportPage, outputPath string) (buildError error) {
	var destination *responses.FPDF_CreateNewDocument
	if destination, buildError = p.Instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{}); buildError != nil {
		p.Log.Error(buildError, "Could not create new document")
		return
	}

	var source *responses.FPDF_LoadDocument
	copyPages := func(sourceFile *string, pageFrom, pageTo, insertIndex int) {
		if buildError != nil {
			return
		}
		if source, buildError = p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
			Path: sourceFile,
		}); buildError != nil {
			p.Log.Error(buildError, "Could not open source PDF", "file_name", sourceFile)
			return
		}
		pageRange := fmt.Sprintf("%d-%d", pageFrom, pageTo)
		p.Log.V(1).Info("Adding pages", "page_range", pageRange, "index", insertIndex, "filename", sourceFile)
		if _, buildError = p.Instance.FPDF_ImportPages(&requests.FPDF_ImportPages{
			Source:      source.Document,
			Destination: destination.Document,
			PageRange:   &pageRange,
			Index:       insertIndex,
		}); buildError != nil {
			p.Log.Error(buildError, "Could not import pages", "file_name", sourceFile)
			return
		}
	}

	pageCount := 0
	bookmarks := make([]pdfcpu.Bookmark, 0, len(episodes))
	for _, episode := range episodes {
		copyPages(&episode.Filename, episode.PageFrom, episode.PageTo, pageCount)
		if len(episode.Title) > 0 {
			bookmarks = append(bookmarks, pdfcpu.Bookmark{
				Title:    episode.Title,
				PageFrom: pageCount + 1,
				PageThru: pageCount + (episode.PageTo - episode.PageFrom) + 1,
			})
		}
		pageCount += (episode.PageTo - episode.PageFrom) + 1
	}

	if buildError != nil {
		p.Log.Error(buildError, "Could not copy all pages to document", "file_name", outputPath)
		return
	}

	saveAsCopy, buildError := p.Instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    requests.SaveFlagIncremental,
		Document: destination.Document,
		FilePath: &outputPath,
	})
	if buildError != nil {
		p.Log.Error(buildError, "Could not save document", "file_name", outputPath)
		return
	}

	p.Log.Info("Attempting to add bookmarks", "filepath", *saveAsCopy.FilePath)
	buildError = pdfApi.AddBookmarksFile(*saveAsCopy.FilePath, *saveAsCopy.FilePath, bookmarks, true, nil)

	return
}

func (p *Reader) Credits(filename string, startPage int, endPage int) (credits string, err error) {
	source, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: &filename,
	})

	defer func() {
		_, err := p.Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: source.Document})
		if err != nil {
			p.Log.Info("Could not close source", "file_name", filename)
			return
		}
	}()

	if err != nil {
		p.Log.Error(err, "Could not open file")
		return "", err
	}

	p.Log.V(1).Info(fmt.Sprintf("Reading %s", filename))
	var creditTypes = []string{"script", "art", "colours", "letters"}
	var textPage *responses.FPDFText_LoadPage
	var scriptRect *responses.FPDFText_GetRect

	for pageIndex := startPage; pageIndex <= endPage; pageIndex++ {
		p.Log.V(1).Info(fmt.Sprintf("Scanning page %d of %s", pageIndex, filename))
		if pdfPage, err := p.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: source.Document,
			Index:    pageIndex - 1,
		}); err != nil {
			p.Log.Error(err, fmt.Sprintf("Failed to load page %d", pageIndex))
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

func (p *Reader) findScriptRect(pageRef references.FPDF_PAGE) (*responses.FPDFText_LoadPage, *responses.FPDFText_GetRect) {
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
			p.Log.V(1).Info(fmt.Sprintf("Found script at %+v", rect))
			scriptRect = rect
		}
	}
	return textPage, scriptRect
}
