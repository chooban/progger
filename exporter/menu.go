package exporter

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/api"
)

func MainMenu(app *api.ProggerApp) *fyne.MainMenu {
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("Progger", fyne.NewMenuItem("Settings...", func() {
		app.AppContext.ShowSettings.Set(true)
	})))

	return mainMenu
}
