package prefs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

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

type Prefs struct {
	RebellionPassword string
	RebellionUsername string
}
