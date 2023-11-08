//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/chooban/progdl-go/internal/pdfium"
	"os"
)

func main() {
	parser := argparse.NewParser("export", "Exports a PDF of selected series")

	series := parser.String("s", "series", &argparse.Options{Required: true, Help: "Series name"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	appEnv := env.NewAppEnv()
	appEnv.Db = db.Init("progs.db")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	var episodes []db.Episode

	appEnv.Db.Preload("Issue").Table("episodes e").
		Joins("join series s on s.id = e.series_id").
		Joins("join issues i on e.issue_id = i.id").
		Where("s.title = ? and e.issue_id > 0", series).
		Order("e.title, part ASC").
		Find(&episodes)

	appEnv.Pdf.Build(episodes)
}
