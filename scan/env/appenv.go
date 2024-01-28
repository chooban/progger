package env

import (
	"github.com/rs/zerolog"
)

type ToSkip struct {
	SeriesTitles []string
}

type AppEnv struct {
	//Db    *gorm.DB
	Log *zerolog.Logger
	//Pdf   pdf.Reader
	Skip  ToSkip
	Known ToSkip
}

func NewAppEnv() AppEnv {
	appEnv := AppEnv{
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
