package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/prefs"
)

type cb func()

func NewSettingsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	fyneApp := a.FyneApp

	allSettings := container.New(
		layout.NewVBoxLayout(),
		directoriesContainer(fyneApp, a.RootWindow),
		widget.NewSeparator(),
		rebellionContainer(fyneApp),
	)

	return allSettings
}

func rebellionContainer(a fyne.App) *fyne.Container {
	pass := widget.NewPasswordEntry()
	pass.Bind(prefs.BoundRebellionPassword(a))

	formContainer := container.New(
		layout.NewVBoxLayout(),
		widget.NewLabel("Rebellion Account"),
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("Username"),
			widget.NewEntryWithData(prefs.BoundRebellionUsername(a)),
			widget.NewLabel("Password"),
			pass,
		),
	)

	return formContainer
}

func directoriesContainer(a fyne.App, w fyne.Window) *fyne.Container {
	boundSource := prefs.BoundSourceDir(a)
	boundExport := prefs.BoundExportDir(a)

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
