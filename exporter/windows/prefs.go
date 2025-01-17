package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/services"
)

func newSettingsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	fyneApp := a.FyneApp

	allSettings := container.New(
		layout.NewVBoxLayout(),
		directoriesContainer(a, a.RootWindow),
		widget.NewSeparator(),
		rebellionContainer(fyneApp),
	)

	return allSettings
}

func rebellionContainer(a fyne.App) *fyne.Container {
	pass := widget.NewPasswordEntry()
	pass.Bind(services.BoundRebellionPassword(a))

	formContainer := container.New(
		layout.NewVBoxLayout(),
		widget.NewLabel("Rebellion Account"),
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("Username"),
			widget.NewEntryWithData(services.BoundRebellionUsername(a)),
			widget.NewLabel("Password"),
			pass,
		),
	)

	return formContainer
}

func directoriesContainer(a *app.ProggerApp, w fyne.Window) *fyne.Container {
	progSource := a.AppService.Prefs.ProgSourceDir
	megSource := a.AppService.Prefs.MegazineSourceDir
	boundExport := a.AppService.Prefs.BoundExportDir

	directoriesFormContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabelWithData(progSource),
		widget.NewButton("Choose Prog Input Directory", func() {
			dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
				if l != nil {
					progSource.Set(l.Path())
				}
			}, w)
		}),
		widget.NewLabelWithData(megSource),
		widget.NewButton("Choose Megazine Input Directory", func() {
			dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
				if l != nil {
					megSource.Set(l.Path())
				}
			}, w)
		}),
		widget.NewLabelWithData(boundExport),
		widget.NewButton("Choose Export Directory", func() {
			dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
				if l != nil {
					boundExport.Set(l.Path())
				}
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
