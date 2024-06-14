package services

import (
	"fyne.io/fyne/v2"
)

type AppServices struct {
	Downloader *Downloader
	Exporter   *Exporter
	Scanner    *Scanner
	Prefs      *Prefs
}

func NewAppServices(a fyne.App) *AppServices {

	return &AppServices{
		Downloader: NewDownloader(),
		Exporter:   NewExporter(),
		Scanner:    NewScanner(),
		Prefs:      NewPrefs(a),
	}
}
