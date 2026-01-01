package app

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/services"
)

type AppServices struct {
	Downloader *services.Downloader
	Exporter   *services.Exporter
	Scanner    *services.Scanner
	Prefs      *Prefs
	Storage    *services.Storage
}

func NewAppServices(a fyne.App) *AppServices {

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic("could not get user config dir")
	}
	proggerConfigDir := filepath.Join(configDir, "progger")
	_, err = os.Stat(proggerConfigDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(proggerConfigDir, 0755)
		if err != nil {
			println(err.Error())
			panic("could not create progger config dir")
		}
	}

	storage := services.NewStorage(proggerConfigDir)

	return &AppServices{
		Downloader: services.NewDownloader(configDir, storage),
		Exporter:   services.NewExporter(),
		Scanner:    services.NewScanner(storage),
		Prefs:      NewPrefs(a),
		Storage:    storage,
	}
}
