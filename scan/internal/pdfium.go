package internal

import (
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

func (r *Reader) Open(filename string) (*FileReader, error) {
	source, err := r.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{Path: &filename})
	if err != nil {
		return nil, err
	}
	return &FileReader{
		Log:      r.Log,
		Instance: r.Instance,
		Doc:      source.Document,
		filename: filename,
	}, nil
}

type FileReader struct {
	Log      logr.Logger
	Instance pdfium.Pdfium
	Doc      references.FPDF_DOCUMENT
	filename string
}

func (fr *FileReader) Close() {
	fr.Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: fr.Doc})
}

func (fr *FileReader) PageCount() (int, error) {
	pageCount, err := fr.Instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: fr.Doc})
	if err != nil {
		return 0, err
	}
	return pageCount.PageCount, nil
}

func (fr *FileReader) Bookmarks() ([]EpisodeDetails, error) {
	f, err := os.Open(fr.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bookmarks, err := pdfApi.Bookmarks(f, model.NewDefaultConfiguration())
	if err != nil {
		return nil, err
	}

	details := make([]EpisodeDetails, len(bookmarks))
	for i, v := range bookmarks {
		details[i] = EpisodeDetails{
			Bookmark: PdfBookmark{
				Title:    v.Title,
				PageFrom: v.PageFrom,
				PageThru: v.PageThru,
			},
		}
	}
	lastIdx := len(details) - 1
	if lastIdx >= 0 && details[lastIdx].Bookmark.PageThru == 0 {
		if pageCount, err := fr.PageCount(); err == nil {
			details[lastIdx].Bookmark.PageThru = pageCount
		}
	}
	return details, nil
}

func (fr *FileReader) Text(pageNumber int) (string, error) {
	pdfPage, err := fr.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
		Document: fr.Doc,
		Index:    pageNumber,
	})
	if err != nil {
		return "", err
	}
	defer fr.Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: pdfPage.Page})

	text, err := fr.Instance.GetPageText(&requests.GetPageText{
		Page: requests.Page{ByReference: &pdfPage.Page},
	})
	if err != nil {
		return "", err
	}
	return text.Text, nil
}

func (fr *FileReader) Credits(startPage, endPage int) (string, error) {
	fr.Log.V(1).Info(fmt.Sprintf("Reading credits from %s (pages %d-%d)", fr.filename, startPage, endPage))

	var creditTypes = []string{"script", "art", "colours", "letters"}
	var textPage *responses.FPDFText_LoadPage
	var scriptRect *responses.FPDFText_GetRect

	for pageIndex := startPage; pageIndex <= endPage; pageIndex++ {
		fr.Log.V(1).Info(fmt.Sprintf("Scanning page %d of %s", pageIndex, fr.filename))
		if pdfPage, err := fr.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: fr.Doc,
			Index:    pageIndex - 1,
		}); err != nil {
			fr.Log.Error(err, fmt.Sprintf("Failed to load page %d", pageIndex))
			return "", errors.New("failed to load page")
		} else {
			if textPage, scriptRect = findScriptRect(fr.Instance, fr.Log, pdfPage.Page); scriptRect != nil {
				fr.Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: pdfPage.Page})
				break
			}
			fr.Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: pdfPage.Page})
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
		creditsText, _ := fr.Instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
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
	if earliestIdx > latestIdx {
		return "", errors.New("invalid credits")
	}
	tmpCredits := tokenized[earliestIdx:min(latestIdx+4, len(tokenized))]
	credits := strings.Join(tmpCredits, " ")

	return credits, nil
}

func (fr *FileReader) TrimTrailingAdverts(pageFrom, pageTo int) int {
	return TrimAdvertPages(pageFrom, pageTo, func(pageIndex int) bool {
		text, err := fr.Text(pageIndex - 1)
		if err != nil {
			return false
		}
		return IsAdvertPageText(text)
	})
}

func (r *Reader) Bookmarks(filename string) ([]EpisodeDetails, error) {
	fr, err := r.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	return fr.Bookmarks()
}

func (r *Reader) Credits(filename string, startPage, endPage int) (string, error) {
	fr, err := r.Open(filename)
	if err != nil {
		return "", err
	}
	defer fr.Close()
	return fr.Credits(startPage, endPage)
}

func (r *Reader) Text(filename string, pageNumber int) (string, error) {
	fr, err := r.Open(filename)
	if err != nil {
		return "", err
	}
	defer fr.Close()
	return fr.Text(pageNumber)
}

func (r *Reader) PageCount(filename string) (int, error) {
	fr, err := r.Open(filename)
	if err != nil {
		return 0, err
	}
	defer fr.Close()
	return fr.PageCount()
}

func findScriptRect(instance pdfium.Pdfium, log logr.Logger, pageRef references.FPDF_PAGE) (*responses.FPDFText_LoadPage, *responses.FPDFText_GetRect) {
	var (
		textPage   *responses.FPDFText_LoadPage
		scriptRect *responses.FPDFText_GetRect
	)
	textPage, _ = instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{
		Page: requests.Page{
			ByReference: &pageRef,
			ByIndex:     nil,
		},
	})
	rects, _ := instance.FPDFText_CountRects(&requests.FPDFText_CountRects{
		TextPage:   textPage.TextPage,
		StartIndex: 0,
		Count:      -1,
	})
	for textRectIndex := 0; textRectIndex < rects.Count; textRectIndex++ {
		rect, _ := instance.FPDFText_GetRect(&requests.FPDFText_GetRect{
			TextPage: textPage.TextPage,
			Index:    textRectIndex,
		})
		text, _ := instance.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
			TextPage: textPage.TextPage,
			Left:     rect.Left,
			Top:      rect.Top,
			Right:    rect.Right,
			Bottom:   rect.Bottom,
		})
		if strings.ToLower(text.Text) == "script" {
			log.V(1).Info(fmt.Sprintf("Found script at %+v", rect))
			scriptRect = rect
		}
	}
	return textPage, scriptRect
}
