package services

import (
	"fyne.io/fyne/v2"
	"os"
	"path/filepath"
)

type AppServices struct {
	Downloader *Downloader
	Exporter   *Exporter
	Scanner    *Scanner
	Prefs      *Prefs
	Storage    *Storage
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

	storage := NewStorage(proggerConfigDir)

	return &AppServices{
		Downloader: NewDownloader(configDir, storage),
		Exporter:   NewExporter(),
		Scanner:    NewScanner(storage),
		Prefs:      NewPrefs(a),
		Storage:    storage,
	}
}
