package pdfium

import (
	"fmt"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
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
