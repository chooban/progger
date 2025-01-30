package internal

import (
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/enums"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/klippa-app/go-pdfium/structs"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"strings"
)

type PdfBuilder struct {
	BuildError  error
	instance    pdfium.Pdfium
	destination *responses.FPDF_CreateNewDocument
	savedAs     *responses.FPDF_SaveAsCopy
}

func NewPdfBuilder() *PdfBuilder {
	return &PdfBuilder{
		instance: Instance,
	}
}

func (p *PdfBuilder) OpenDestination() {
	p.destination, p.BuildError = p.instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
}

func (p *PdfBuilder) CopyStrippedPages(sourceFile *string, pageFrom, pageTo, insertIndex int) (pageCount int) {
	if p.BuildError != nil {
		return 0
	}
	var source *responses.FPDF_LoadDocument
	if source, p.BuildError = p.instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: sourceFile,
	}); p.BuildError != nil {
		return
	}
	for pageNum := pageFrom - 1; pageNum < pageTo; pageNum++ {
		if p.shouldSkipPage(source, pageNum) {
			continue
		}
		ref, err := p.instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: source.Document,
			Index:    pageNum,
		})
		if err != nil {
			p.BuildError = err
			return
		}
		width, _ := p.instance.FPDF_GetPageWidth(&requests.FPDF_GetPageWidth{requests.Page{ByReference: &ref.Page}})
		height, _ := p.instance.FPDF_GetPageHeight(&requests.FPDF_GetPageHeight{requests.Page{ByReference: &ref.Page}})

		var res *responses.FPDFPage_CountObjects
		if res, err = p.instance.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{
			requests.Page{
				ByReference: &ref.Page,
				ByIndex:     nil,
			},
		}); err != nil {
			p.BuildError = err
			return
		}

		var bgObject *responses.FPDFPage_GetObject
		for i := 0; i < res.Count; i++ {
			obj, _ := p.instance.FPDFPage_GetObject(&requests.FPDFPage_GetObject{
				Page:  requests.Page{ByReference: &ref.Page},
				Index: i,
			})
			t, _ := p.instance.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: obj.PageObject})
			if t.Type != enums.FPDF_PAGEOBJ_IMAGE {
				continue
			} else {
				bgObject = obj
				break
			}
		}
		if bgObject == nil {
			p.BuildError = fmt.Errorf("pdf_page_object not found")
			return
		}

		// This is the raw, compressed image. We want to then embed that in a new page somehow
		rawImage, _ := p.instance.FPDFImageObj_GetImageDataRaw(&requests.FPDFImageObj_GetImageDataRaw{
			bgObject.PageObject,
		})
		newPage, _ := p.instance.FPDFPage_New(&requests.FPDFPage_New{
			Document:  p.destination.Document,
			PageIndex: insertIndex + (pageCount),
			Width:     width.Width,
			Height:    height.Height,
		})
		newImage, _ := p.instance.FPDFPageObj_NewImageObj(&requests.FPDFPageObj_NewImageObj{
			p.destination.Document,
		})
		if _, err := p.instance.FPDFImageObj_LoadJpegFileInline(&requests.FPDFImageObj_LoadJpegFileInline{
			Page: &requests.Page{
				ByReference: &newPage.Page,
			},
			ImageObject: newImage.PageObject,
			FileData:    rawImage.Data,
			Count:       0,
		}); err != nil {
			p.BuildError = err
			return
		}
		meta, _ := p.instance.FPDFImageObj_GetImageMetadata(&requests.FPDFImageObj_GetImageMetadata{
			ImageObject: bgObject.PageObject,
			Page: requests.Page{
				ByReference: &ref.Page,
			},
		})

		p.instance.FPDFPageObj_Transform(&requests.FPDFPageObj_Transform{
			PageObject: newImage.PageObject,
			Transform: structs.FPDF_FS_MATRIX{
				float32(meta.ImageMetadata.Width) / 2.7,
				0,
				0,
				float32(meta.ImageMetadata.Height) / 2.7,
				0,
				0,
			},
		})

		p.instance.FPDFPage_InsertObject(&requests.FPDFPage_InsertObject{
			Page: requests.Page{
				ByReference: &newPage.Page,
			},
			PageObject: newImage.PageObject,
		})
		p.instance.FPDFPage_GenerateContent(&requests.FPDFPage_GenerateContent{
			Page: requests.Page{
				ByReference: &newPage.Page,
			},
		})

		p.instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: newPage.Page})

		pageCount++
	}
	return
}

func (p *PdfBuilder) shouldSkipPage(source *responses.FPDF_LoadDocument, pageIndex int) bool {
	ref, err := p.instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: requests.Page{
		ByIndex: &requests.PageByIndex{
			Document: source.Document,
			Index:    pageIndex,
		},
	}})
	if err != nil {
		// Bad page ref?
		println(err.Error())
		return true
	}
	r, err := p.instance.FPDFText_GetText(&requests.FPDFText_GetText{
		TextPage:   ref.TextPage,
		StartIndex: 0,
		Count:      100,
	})
	if err != nil {
		// No text?
		return true
	}
	if strings.Contains(strings.ToLower(r.Text), "on sale now") {
		return true
	}
	return false
}

func (p *PdfBuilder) CopyPages(sourceFile *string, pageFrom, pageTo, insertIndex int) int {
	if p.BuildError != nil {
		return 0
	}
	var source *responses.FPDF_LoadDocument
	if source, p.BuildError = p.instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: sourceFile,
	}); p.BuildError != nil {
		return 0
	}

	// Sometimes the tail end of an episode has adverts. We can try and filter them out
	for pageIndex := pageTo; pageIndex > pageFrom; pageIndex-- {
		if p.shouldSkipPage(source, pageIndex) {
			pageFrom--
			continue
		}

		// If we didn't continue then assume we're into episode pages. Conceivably, the phrase "on sale now" might
		// be in the dialogue, so going through all the pages doesn't make sense.
		break
	}
	pageRange := fmt.Sprintf("%d-%d", pageFrom, pageTo)
	_, p.BuildError = p.instance.FPDF_ImportPages(&requests.FPDF_ImportPages{
		Source:      source.Document,
		Destination: p.destination.Document,
		PageRange:   &pageRange,
		Index:       insertIndex,
	})

	return pageTo - pageFrom
}

func (p *PdfBuilder) Save(outputPath string) {
	if p.BuildError != nil {
		return
	}
	p.savedAs, p.BuildError = p.instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    requests.SaveFlagIncremental,
		Document: p.destination.Document,
		FilePath: &outputPath,
	})
}

func (p *PdfBuilder) AddBookmarks(bookmarks []pdfcpu.Bookmark) {
	if p.BuildError != nil {
		return
	}
	p.BuildError = pdfApi.AddBookmarksFile(*p.savedAs.FilePath, *p.savedAs.FilePath, bookmarks, true, nil)
}

func (p *PdfBuilder) Build(episodes []api.ExportPage, artistsEdition bool, outputPath string) (buildError error) {
	p.OpenDestination()

	pageCount := 0
	bookmarks := make([]pdfcpu.Bookmark, 0, len(episodes))
	for _, episode := range episodes {
		pagesAdded := 0
		if artistsEdition {
			pagesAdded = p.CopyStrippedPages(&episode.Filename, episode.PageFrom, episode.PageTo, pageCount)
		} else {
			pagesAdded = p.CopyPages(&episode.Filename, episode.PageFrom, episode.PageTo, pageCount)
		}
		if len(episode.Title) > 0 {
			bookmarks = append(bookmarks, pdfcpu.Bookmark{
				Title:    episode.Title,
				PageFrom: pageCount + 1,
				PageThru: pageCount + pagesAdded + 1,
			})
		}
		pageCount += pagesAdded + 1
	}
	p.Save(outputPath)
	p.AddBookmarks(bookmarks)

	return p.BuildError
}
