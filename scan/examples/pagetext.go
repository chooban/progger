//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan/internal"
	"github.com/go-logr/zerologr"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func main() {

	parser := argparse.NewParser("pagetext", "Try to print text on PDF page")

	filename := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to parse"})
	page := parser.Int("p", "page", &argparse.Options{Required: true, Help: "Page to inspect"})

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

	//ctx := context.Background()
	//ctx = logr.NewContext(ctx, log)

	p := internal.NewPdfiumReader(log)
	contents, err := os.ReadFile(*filename)
	doc, err := p.Instance.OpenDocument(&requests.OpenDocument{
		File: &contents,
	})
	if err != nil {
		println(err.Error())
		return
	}
	ref, _ := p.Instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: requests.Page{
		ByIndex: &requests.PageByIndex{
			Document: doc.Document,
			Index:    *page,
		},
	}})
	r, err := p.Instance.FPDFText_GetText(&requests.FPDFText_GetText{
		TextPage:   ref.TextPage,
		StartIndex: 0,
		Count:      100,
	})
	if err != nil {
		println(err.Error())
		return
	}

	println(r.Text)
}
