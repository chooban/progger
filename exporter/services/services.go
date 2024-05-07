package services

import (
	"context"
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/prefs"
)

type AppServices struct {
	Downloader *Downloader
	Exporter   *Exporter
	Scanner    *Scanner
}

func NewAppServices(ctx context.Context, a fyne.App, p *prefs.Prefs) *AppServices {
	// We want to be able to react to the source directory changing
	boundSource := prefs.BoundSourceDir(a)
	boundExport := prefs.BoundExportDir(a)

	return &AppServices{
		Downloader: NewDownloader(ctx, boundSource),
		Exporter:   NewExporter(ctx, boundSource, boundExport),
		Scanner:    NewScanner(ctx),
	}
}
