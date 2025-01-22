//go:build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/klippa-app/go-pdfium/enums"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/structs"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func main() {

	parser := argparse.NewParser("pagetext", "Try to print text on PDF page")

	filename := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to parse"})
	//page := parser.Int("p", "page", &argparse.Options{Required: true, Help: "Page to inspect"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(writer)
	logger = logger.With().Caller().Timestamp().Logger()
	var log = zerologr.New(&logger)

	p := pdfium.NewPdfiumReader(log)
	doc, err := p.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: filename,
	})
	if err != nil {
		log.Error(err, "Could not open document")
		return
	}

	pageCount, err := p.Instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		log.Info("Could not get page count")
		return
	}
	newFile, _ := p.Instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})

	for i := 0; i < pageCount.PageCount; i++ {
		getBackgroundAsPage(p, log, doc.Document, &i, newFile.Document)
	}

	log.Info("Writing")
	outputPath := "./artistsedition.pdf"
	if asCopy, err := p.Instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Document: newFile.Document,
		FilePath: &outputPath,
	}); err != nil {
		log.Error(err, "Could not save as copy")
		return
	} else {
		p.Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: newFile.Document})
		log.Info("Saved file", "path", asCopy.FilePath)
	}
}

func getBackgroundAsPage(p *pdfium.Reader, log logr.Logger, doc references.FPDF_DOCUMENT, page *int, newFile references.FPDF_DOCUMENT) {
	ref, _ := p.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
		Document: doc,
		Index:    *page,
	})
	width, _ := p.Instance.FPDF_GetPageWidth(&requests.FPDF_GetPageWidth{requests.Page{ByReference: &ref.Page}})
	height, _ := p.Instance.FPDF_GetPageHeight(&requests.FPDF_GetPageHeight{requests.Page{ByReference: &ref.Page}})
	log.Info(fmt.Sprintf("Page width/height is %0.2f/%0.2f", width.Width, height.Height))

	res, err := p.Instance.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{
		requests.Page{
			ByReference: &ref.Page,
			ByIndex:     nil,
		},
	})
	if err != nil {
		log.Error(err, "Could not count objects")
		return
	}

	for i := 0; i < res.Count; i++ {
		obj, _ := p.Instance.FPDFPage_GetObject(&requests.FPDFPage_GetObject{
			Page:  requests.Page{ByReference: &ref.Page},
			Index: i,
		})
		t, _ := p.Instance.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: obj.PageObject})
		if t.Type != enums.FPDF_PAGEOBJ_IMAGE {
			continue
		}
		meta, _ := p.Instance.FPDFImageObj_GetImageMetadata(&requests.FPDFImageObj_GetImageMetadata{
			ImageObject: obj.PageObject,
			Page: requests.Page{
				ByReference: &ref.Page,
			},
		})

		//log.Info(fmt.Sprintf("Image metadata is %+v", meta))

		//bm, _ := p.Instance.FPDFImageObj_GetBitmap(&requests.FPDFImageObj_GetBitmap{
		//	obj.PageObject,
		//})
		// This is the raw, compressed image. We want to then embed that in a new page somehow
		rawImage, _ := p.Instance.FPDFImageObj_GetImageDataRaw(&requests.FPDFImageObj_GetImageDataRaw{
			obj.PageObject,
		})
		newPage, _ := p.Instance.FPDFPage_New(&requests.FPDFPage_New{
			Document:  newFile,
			PageIndex: *page,
			Width:     width.Width,
			Height:    height.Height,
		})
		newImage, _ := p.Instance.FPDFPageObj_NewImageObj(&requests.FPDFPageObj_NewImageObj{
			newFile,
		})
		p.Instance.FPDFImageObj_LoadJpegFileInline(&requests.FPDFImageObj_LoadJpegFileInline{
			Page: &requests.Page{
				ByReference: &newPage.Page,
			},
			ImageObject: newImage.PageObject,
			FileData:    rawImage.Data,
			Count:       0,
		})
		//https://groups.google.com/g/pdfium/c/z0hpHKVYIEY
		//p.Instance.FPDFImageObj_SetBitmap(
		//	&requests.FPDFImageObj_SetBitmap{&requests.Page{ByReference: &newPage.Page},
		//		1,
		//		newImage.PageObject,
		//		bm.Bitmap,
		//	},
		//)
		p.Instance.FPDFPageObj_Transform(&requests.FPDFPageObj_Transform{
			PageObject: newImage.PageObject,
			Transform: structs.FPDF_FS_MATRIX{
				float32(meta.ImageMetadata.Width) / 2.8,
				0,
				0,
				float32(meta.ImageMetadata.Height) / 2.8,
				0,
				0,
			},
		})

		p.Instance.FPDFPage_InsertObject(&requests.FPDFPage_InsertObject{
			Page: requests.Page{
				ByReference: &newPage.Page,
			},
			PageObject: newImage.PageObject,
		})
		p.Instance.FPDFPage_GenerateContent(&requests.FPDFPage_GenerateContent{
			Page: requests.Page{
				ByReference: &newPage.Page,
			},
		})

		p.Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: newPage.Page})

		break
	}

}
