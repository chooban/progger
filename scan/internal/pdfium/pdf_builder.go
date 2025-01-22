package pdfium

import (
	"fmt"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/enums"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/klippa-app/go-pdfium/structs"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type pdfBuilder struct {
	BuildError  error
	instance    pdfium.Pdfium
	destination *responses.FPDF_CreateNewDocument
	savedAs     *responses.FPDF_SaveAsCopy
}

func (p *pdfBuilder) OpenDestination() {
	p.destination, p.BuildError = p.instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
}

func (p *pdfBuilder) CopyStrippedPages(sourceFile *string, pageFrom, pageTo, insertIndex int) {
	if p.BuildError != nil {
		return
	}
	var source *responses.FPDF_LoadDocument
	if source, p.BuildError = p.instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: sourceFile,
	}); p.BuildError != nil {
		return
	}
	for pageNum := pageFrom; pageNum <= pageTo; pageNum++ {
		ref, err := p.instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
			Document: source.Document,
			Index:    pageNum - 1,
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
		//pageIndex := insertIndex + (pageFrom - pageNum)
		newPage, _ := p.instance.FPDFPage_New(&requests.FPDFPage_New{
			Document:  p.destination.Document,
			PageIndex: insertIndex + (pageFrom + pageNum),
			Width:     width.Width,
			Height:    height.Height,
		})
		newImage, _ := p.instance.FPDFPageObj_NewImageObj(&requests.FPDFPageObj_NewImageObj{
			p.destination.Document,
		})
		p.instance.FPDFImageObj_LoadJpegFileInline(&requests.FPDFImageObj_LoadJpegFileInline{
			Page: &requests.Page{
				ByReference: &newPage.Page,
			},
			ImageObject: newImage.PageObject,
			FileData:    rawImage.Data,
			Count:       0,
		})
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
	}
}

func (p *pdfBuilder) CopyPages(sourceFile *string, pageFrom, pageTo, insertIndex int) {
	if p.BuildError != nil {
		return
	}
	var source *responses.FPDF_LoadDocument
	if source, p.BuildError = p.instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: sourceFile,
	}); p.BuildError != nil {
		return
	}
	pageRange := fmt.Sprintf("%d-%d", pageFrom, pageTo)
	_, p.BuildError = p.instance.FPDF_ImportPages(&requests.FPDF_ImportPages{
		Source:      source.Document,
		Destination: p.destination.Document,
		PageRange:   &pageRange,
		Index:       insertIndex,
	})
}

func (p *pdfBuilder) Save(outputPath string) {
	if p.BuildError != nil {
		return
	}
	p.savedAs, p.BuildError = p.instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    requests.SaveFlagIncremental,
		Document: p.destination.Document,
		FilePath: &outputPath,
	})
}

func (p *pdfBuilder) AddBookmarks(bookmarks []pdfcpu.Bookmark) {
	if p.BuildError != nil {
		return
	}
	p.BuildError = pdfApi.AddBookmarksFile(*p.savedAs.FilePath, *p.savedAs.FilePath, bookmarks, true, nil)
}
