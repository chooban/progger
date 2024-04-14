package prefs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type cb func()

func ShowPrefs(a fyne.App, w fyne.Window, onClose cb) *fyne.Container {

	allSettings := container.New(
		layout.NewVBoxLayout(),
		directoriesContainer(a, w),
		widget.NewSeparator(),
		rebellionContainer(a, w),
	)

	prefsContainer := container.NewBorder(
		container.NewCenter(
			widget.NewLabel("Settings"),
		),
		widget.NewButton("Close", onClose),
		nil,
		nil,
		allSettings,
	)

	return prefsContainer
}

func rebellionContainer(a fyne.App, w fyne.Window) *fyne.Container {
	pass := widget.NewPasswordEntry()
	pass.Bind(BoundRebellionPassword(a))

	formContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Username"),
		widget.NewEntryWithData(BoundRebellionUsername(a)),
		widget.NewLabel("Password"),
		pass,
	)

	return formContainer
}

func directoriesContainer(a fyne.App, w fyne.Window) *fyne.Container {
	boundSource := BoundSourceDir(a)
	boundExport := BoundExportDir(a)

	directoriesFormContainer := container.New(
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

	directoriesContainer := container.New(
		layout.NewVBoxLayout(),
		widget.NewLabel("Directories"),
		directoriesFormContainer,
	)

	return directoriesContainer
}

func boundStringValue(app fyne.App, bindName string) binding.String {
	v := app.Preferences().String(bindName)
	b := binding.NewString()
	b.Set(v)
	b.AddListener(binding.NewDataListener(func() {
		if newV, _ := b.Get(); newV != v {
			app.Preferences().SetString(bindName, newV)
		}
	}))

	return b
}

func BoundSourceDir(app fyne.App) binding.String {
	return boundStringValue(app, "SourceDir")
}

func BoundExportDir(app fyne.App) binding.String {
	return boundStringValue(app, "ExportDir")
}

func BoundRebellionUsername(app fyne.App) binding.String {
	return boundStringValue(app, "RebellionUsername")
}

func BoundRebellionPassword(app fyne.App) binding.String {
	return boundStringValue(app, "RebellionPassword")
}
