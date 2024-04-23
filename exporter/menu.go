package exporter

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
)

func MainMenu(app *api.ProggerApp) *fyne.MainMenu {
	about := fyne.NewMenuItem("About", func() {
		d := app.FyneApp.NewWindow("About")
		d.SetContent(widget.NewLabel("This is about the app"))
		d.Show()
	})
	settings := fyne.NewMenuItem("Settings...", func() {
		app.AppContext.ShowSettings.Set(true)
	})
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("Progger", settings, about))

	return mainMenu
}
