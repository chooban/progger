package services

import (
	"context"
	"fyne.io/fyne/v2"
)

type AppServices struct {
	Downloader *Downloader
	Exporter   *Exporter
	Scanner    *Scanner
	Prefs      *Prefs
}

func NewAppServices(ctx context.Context, a fyne.App) *AppServices {

	return &AppServices{
		Downloader: NewDownloader(ctx),
		Exporter:   NewExporter(ctx),
		Scanner:    NewScanner(ctx),
		Prefs:      NewPrefs(a),
	}
}
