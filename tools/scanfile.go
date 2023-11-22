//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/pdfium"
	"github.com/chooban/progdl-go/internal/scanner"
	"github.com/rs/zerolog"
	"os"
)

func main() {

	parser := argparse.NewParser("scan", "Scans a pdf")
	f := parser.String("f", "directory", &argparse.Options{Required: true, Help: "Directory to scan"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	appEnv := env.NewAppEnv()
	//appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	issue, _ := scanner.ScanFile(appEnv, *f)

	for _, v := range issue.Episodes {
		appEnv.Log.Info().Msg(fmt.Sprintf("Title: %s", v.Title))
		appEnv.Log.Info().Msg(fmt.Sprintf("Writers: %+v", v.Script))
	}
}
