package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/prefs"
)

type cb func()

func NewSettings(a *app.ProggerApp) {
	settingsWindow := a.FyneApp.NewWindow("Settings")
	settingsWindow.SetCloseIntercept(func() {
		a.AppContext.Dispatch(app.HideSettingsMessage{})
	})
	settingsWindow.SetContent(ShowPrefs(a.FyneApp, settingsWindow, func() {
		a.AppContext.Dispatch(app.HideSettingsMessage{})
	}))
	settingsWindow.Resize(fyne.NewSquareSize(600))

	a.AppContext.ShowSettings.AddListener(binding.NewDataListener(func() {
		if showSettings, _ := a.AppContext.ShowSettings.Get(); showSettings {
			settingsWindow.Show()
		} else {
			settingsWindow.Hide()
		}
	}))
}

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
	pass.Bind(prefs.BoundRebellionPassword(a))

	formContainer := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Username"),
		widget.NewEntryWithData(prefs.BoundRebellionUsername(a)),
		widget.NewLabel("Password"),
		pass,
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
