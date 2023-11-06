//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/klippa-app/go-pdfium/requests"
	"os"
	"strings"
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

	appEnv.Log.Info().Msg(fmt.Sprintf("Loading page %d", *page))

	source, err := appEnv.Pdfium.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
		Path: filename,
	})
	pdfPage, err := appEnv.Pdfium.FPDF_LoadPage(&requests.FPDF_LoadPage{
		Document: source.Document,
		Index:    *page - 1,
	})
	if err != nil {
		appEnv.Log.Err(err).Msg("Failed to load page")
		panic(1)
	}
	textPage, err := appEnv.Pdfium.FPDFText_LoadPage(&requests.FPDFText_LoadPage{
		Page: requests.Page{
			ByIndex:     nil,
			ByReference: &pdfPage.Page,
		}})

	counts, err := appEnv.Pdfium.FPDFText_CountRects(&requests.FPDFText_CountRects{
		TextPage:   textPage.TextPage,
		StartIndex: 0,
		Count:      -1,
	})

	for i := 0; i < counts.Count; i++ {
		rect, _ := appEnv.Pdfium.FPDFText_GetRect(&requests.FPDFText_GetRect{
			TextPage: textPage.TextPage,
			Index:    i,
		})
		text, _ := appEnv.Pdfium.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
			TextPage: textPage.TextPage,
			Left:     rect.Left,
			Top:      rect.Top,
			Right:    rect.Right,
			Bottom:   rect.Bottom,
		})
		if strings.ToLower(text.Text) == "script" {
			appEnv.Log.Info().Msg("Found the script box")
			height := rect.Bottom - rect.Top
			width := rect.Right - rect.Left
			creditsBox, _ := appEnv.Pdfium.FPDFText_GetBoundedText(&requests.FPDFText_GetBoundedText{
				TextPage: textPage.TextPage,
				Left:     rect.Left - (width / 2),
				Top:      rect.Top,
				Right:    rect.Right + width + width/2,
				Bottom:   rect.Bottom + (18 * height),
			})
			appEnv.Log.Info().Msg(creditsBox.Text)
			appEnv.Log.Info().Msg(strings.ReplaceAll(creditsBox.Text, "\r\n", " "))
		}
	}

}
