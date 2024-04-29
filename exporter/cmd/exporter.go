package main

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/windows"
)

func main() {
	app := api.NewProggerApp()

	app.RootWindow.Resize(fyne.NewSize(600, 400))
	app.RootWindow.SetMaster()
	app.RootWindow.SetContent(windows.MainWindow(app))
	app.RootWindow.SetMainMenu(exporter.MainMenu(app))

	windows.NewSettings(app)

	app.RootWindow.ShowAndRun()
}
