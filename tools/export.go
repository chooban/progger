//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/internal/db"
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

	myDb := db.Init("progs.db")

	var episodes []db.Episode

	myDb.Preload("Issue").Table("episodes e").
		Joins("join series s on s.id = e.series_id").
		Joins("join issues i on e.issue_id = i.id").
		Where("s.title = ? and e.issue_id > 0", series).
		Order("e.title, part ASC").
		Find(&episodes)

	//appEnv.Pdf.Build(episodes)
}
