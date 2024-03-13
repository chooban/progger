package exporter

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

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
