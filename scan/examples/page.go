//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func main() {
	parser := argparse.NewParser("scan", "Scans a pdf")
	file := parser.String("f", "file", &argparse.Options{Required: true, Help: "File to scan"})
	pageNumber := parser.Int("p", "page", &argparse.Options{Required: true, Help: "Page to export"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	var log = zerologr.New(&logger)

	ctx := logr.NewContext(context.Background(), log)
	pageToExport := api.ExportPage{
		Filename: *file,
		PageFrom: *pageNumber,
		PageTo:   *pageNumber,
		Title:    "An Example Title",
	}

	if pageBytes, err := scan.BuildPageAsPDF(ctx, pageToExport, true); err != nil {
		log.Error(err, "Failed to export")
	} else {
		createFile("page.pdf")
		var f, err = os.OpenFile("page.pdf", os.O_RDWR, 0644)
		if err != nil {
			log.Error(err, "Failed to write file")
			return
		}
		defer f.Close()
		_, err = f.Write(*pageBytes)
		if err != nil {
			log.Error(err, "Failed to write file")
		}
	}

}
func createFile(path string) {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		checkError(err)
		defer file.Close()
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}
