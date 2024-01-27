//go:build tools
// +build tools

package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/env"
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

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	issue, _ := scan.File(appEnv, *f)

	for _, v := range issue.Episodes {
		appEnv.Log.Info().Msg(fmt.Sprintf("Series: %s", v.Series))
		appEnv.Log.Info().Msg(fmt.Sprintf("Title: %s", v.Title))
		appEnv.Log.Info().Msg(fmt.Sprintf("Writers: %+v", v.Credits))
	}
}
