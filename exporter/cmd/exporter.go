package main

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/windows"
)

func main() {
	a := app.NewProggerApp()

	a.RootWindow.Resize(fyne.NewSize(600, 400))
	a.RootWindow.SetMaster()
	a.RootWindow.SetContent(windows.MainWindow(a))
	a.RootWindow.SetMainMenu(exporter.MainMenu(a))

	windows.NewSettings(a)

	a.RootWindow.ShowAndRun()
}
