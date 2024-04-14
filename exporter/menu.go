package exporter

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/prefs"
)

func MainMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	mainMenu := fyne.NewMainMenu(fyne.NewMenu("Progger", fyne.NewMenuItem("Settings...", func() {
		prefs := prefs.ShowPrefs(a, w, func() {
			w.SetContent(MainWindow(a, w))
		})

		w.SetContent(prefs)
	})))

	return mainMenu
}
