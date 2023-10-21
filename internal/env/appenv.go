package env

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type AppEnv struct {
	Db  *gorm.DB
	Log *zerolog.Logger
}
