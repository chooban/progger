package exporter

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/windows"
)

func MainMenu(a *app.ProggerApp) *fyne.MainMenu {
	about := fyne.NewMenuItem("About", func() {
		d := windows.NewAbout(a)
		d.Show()
	})
	settings := fyne.NewMenuItem("Settings...", func() {
		a.AppContext.Dispatch(app.ShowSettingsMessage{})
	})
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("Progger", settings, about))

	return mainMenu
}
