package env

import (
	"github.com/klippa-app/go-pdfium"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"os"
	"time"
)

type ToSkip struct {
	SeriesTitles []string
}

type AppEnv struct {
	Db     *gorm.DB
	Log    *zerolog.Logger
	Pdfium pdfium.Pdfium
	Skip   ToSkip
	Known  ToSkip
}

func NewAppEnv() AppEnv {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(writer)

	appEnv := AppEnv{
		Db:     nil,
		Log:    &logger,
		Pdfium: instance,
		Skip: ToSkip{
			SeriesTitles: []string{
				"Interrogation",
				"New Books",
				"Obituary",
				"Tribute",
				"Untitled",
			},
		},
		Known: ToSkip{
			SeriesTitles: []string{
				"Anderson, Psi-Division",
				"Strontium Dug",
			},
		},
	}

	return appEnv
}
