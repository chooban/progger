package env

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type ToSkip struct {
	SeriesTitles []string
}

type AppEnv struct {
	Db    *gorm.DB
	Log   *zerolog.Logger
	Skip  ToSkip
	Known ToSkip
}
