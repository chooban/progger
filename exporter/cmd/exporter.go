package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/chooban/progger/exporter"
)

func main() {
	a := app.NewWithID("com.rosshendry.progger.exporter")
	w := a.NewWindow("Progger - Exporter")
	w.Resize(fyne.NewSize(600, 400))

	w.SetContent(exporter.MainWindow(a, w))

	w.ShowAndRun()
}
