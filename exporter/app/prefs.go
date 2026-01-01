package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"github.com/zalando/go-keyring"
)

func BoundRebellionUsername(app fyne.App) binding.String {
	return boundSecretValue(app, "RebellionUsername")
}

func BoundRebellionPassword(app fyne.App) binding.String {
	return boundSecretValue(app, "RebellionPassword")
}

func boundSecretValue(app fyne.App, bindName string) binding.String {
	secret, _ := keyring.Get(app.UniqueID(), bindName)
	b := binding.NewString()
	b.Set(secret)
	b.AddListener(binding.NewDataListener(func() {
		if newV, _ := b.Get(); newV != secret {
			err := keyring.Set(app.UniqueID(), bindName, newV)
			if err != nil {
				println(err.Error())
			}
		}
	}))

	return b
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
	app               fyne.App
	ProgSourceDir     binding.String
	MegazineSourceDir binding.String
	BoundExportDir    binding.String
}

func (p *Prefs) RebellionDetails() (string, string) {
	username, _ := BoundRebellionUsername(p.app).Get()
	password, _ := BoundRebellionPassword(p.app).Get()
	return username, password
}

func (p *Prefs) ProgSourceDirectory() string {
	srcDir, _ := p.ProgSourceDir.Get()

	return srcDir
}

func (p *Prefs) MegSourceDirectory() string {
	srcDir, _ := p.MegazineSourceDir.Get()

	return srcDir
}

func (p *Prefs) ExportDirectory() string {
	exportDir, _ := p.BoundExportDir.Get()

	return exportDir
}

func NewPrefs(a fyne.App) *Prefs {
	return &Prefs{
		app:               a,
		ProgSourceDir:     boundStringValue(a, "ProgSourceDir"),
		MegazineSourceDir: boundStringValue(a, "MegazineSourceDir"),
		BoundExportDir:    boundStringValue(a, "ExportDir"),
	}
}
