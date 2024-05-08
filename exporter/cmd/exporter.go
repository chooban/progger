package main

import (
	"fyne.io/fyne/v2"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/windows"
)

func main() {
	a := app.NewProggerApp()

	a.RootWindow.Resize(fyne.NewSize(800, 1000))
	a.RootWindow.SetMaster()
	a.RootWindow.SetContent(windows.TabWindow(a))

	a.RootWindow.ShowAndRun()
}
