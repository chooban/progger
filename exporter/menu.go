package exporter

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/windows"
)

func MainMenu(app *api.ProggerApp) *fyne.MainMenu {
	about := fyne.NewMenuItem("About", func() {
		d := windows.NewAbout(app)
		d.Show()
	})
	settings := fyne.NewMenuItem("Settings...", func() {
		app.AppContext.Dispatch(api.ShowSettingsMessage{})
	})
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("Progger", settings, about))

	return mainMenu
}
