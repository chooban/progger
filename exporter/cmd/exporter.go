package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/prefs"
)

func main() {
	app := api.NewProggerApp()

	app.RootWindow.Resize(fyne.NewSize(600, 400))
	app.RootWindow.SetMaster()
	app.RootWindow.SetContent(exporter.MainWindow(app))
	app.RootWindow.SetMainMenu(exporter.MainMenu(app))

	settingsWindow := app.FyneApp.NewWindow("Settings")
	settingsWindow.SetContent(prefs.ShowPrefs(app.FyneApp, settingsWindow, func() {
		app.AppContext.ShowSettings.Set(false)
	}))
	settingsWindow.Resize(fyne.NewSquareSize(600))

	app.AppContext.ShowSettings.AddListener(binding.NewDataListener(func() {
		showSettings, _ := app.AppContext.ShowSettings.Get()
		if showSettings {
			settingsWindow.Show()
		} else {
			settingsWindow.Hide()
		}
	}))

	app.RootWindow.ShowAndRun()
}
