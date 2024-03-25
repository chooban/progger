package exporter

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type cb func()

func ShowPrefs(a fyne.App, w fyne.Window, onClose cb) *fyne.Container {
	boundSource := BoundSourceDir(a)
	boundExport := BoundExportDir(a)

	formContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabelWithData(boundSource),
		widget.NewButton("Choose Input Directory", func() {
			dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
				boundSource.Set(l.Path())
			}, w)
		}),
		widget.NewLabelWithData(boundExport),
		widget.NewButton("Choose Export Directory", func() {
			dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
				boundExport.Set(l.Path())
			}, w)
		}),
	)

	prefsContainer := container.NewBorder(
		container.NewCenter(
			widget.NewLabel("Preferences"),
		),
		widget.NewButton("Close", onClose),
		nil,
		nil,
		formContainer,
	)

	return prefsContainer
}

func BoundSourceDir(app fyne.App) binding.String {
	src := app.Preferences().String("SourceDir")
	boundSourceDir := binding.NewString()
	boundSourceDir.Set(src)
	boundSourceDir.AddListener(binding.NewDataListener(func() {
		if newSource, _ := boundSourceDir.Get(); newSource != src {
			println(fmt.Sprintf("Setting source dir to %s", newSource))
			app.Preferences().SetString("SourceDir", newSource)
		}
	}))

	return boundSourceDir
}

func BoundExportDir(app fyne.App) binding.String {
	dest := app.Preferences().String("ExportDir")
	boundSourceDir := binding.NewString()
	boundSourceDir.Set(dest)
	boundSourceDir.AddListener(binding.NewDataListener(func() {
		if newDest, _ := boundSourceDir.Get(); newDest != dest {
			app.Preferences().SetString("ExportDir", newDest)
		}
	}))

	return boundSourceDir
}

func ExportDir(app fyne.App) string {
	dest := app.Preferences().String("ExportDir")
	return dest
}
