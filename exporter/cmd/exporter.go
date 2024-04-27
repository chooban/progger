package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/windows"
)

func main() {
	app := api.NewProggerApp()

	app.RootWindow.Resize(fyne.NewSize(600, 400))
	app.RootWindow.SetMaster()
	app.RootWindow.SetContent(exporter.MainWindow(app))
	app.RootWindow.SetMainMenu(exporter.MainMenu(app))

	app.AppContext.ShowSettings.AddListener(binding.NewDataListener(func() {
		if showSettings, _ := app.AppContext.ShowSettings.Get(); showSettings {
			windows.NewSettings(app)
		}
	}))

	app.RootWindow.ShowAndRun()
}
