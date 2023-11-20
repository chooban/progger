//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"os"
)

func main() {
	parser := argparse.NewParser("pageobjects", "Try to list page objects")

	filename := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to parse"})
	page := parser.Int("p", "page", &argparse.Options{Required: true, Help: "Page to inspect"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	//pdfium := pdfium.Instance
	source, err := pdfium.Instance.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: filename,
	})
	if err != nil {
		//p.Log.Err(err).Msg("Could not open file")
		return
	}
	var pdfPage *responses.FPDF_LoadPage
	if pdfPage, err = pdfium.Instance.FPDF_LoadPage(&requests.FPDF_LoadPage{
		Document: source.Document,
		Index:    *page - 1,
	}); err != nil {
		return
	}
	structuredText, err := pdfium.Instance.GetPageTextStructured(&requests.GetPageTextStructured{
		Page: requests.Page{
			ByReference: &pdfPage.Page,
		},
		Mode:                   requests.GetPageTextStructuredModeBoth,
		CollectFontInformation: false,
		PixelPositions:         requests.GetPageTextStructuredPixelPositions{},
	})

	for _, v := range structuredText.Rects {
		appEnv.Log.Info().Msg(v.Text)
	}
	//credits, err := appEnv.Pdf.Credits(*filename, *page, *page)

	//appEnv.Log.Info().Msg(fmt.Sprintf("Got credits of '%s'", credits))

}
