package services

import (
	"fyne.io/fyne/v2"
	"github.com/zalando/go-keyring"
)

type Prefs struct {
	app fyne.App
}

func (p *Prefs) RebellionDetails() (string, string) {
	username, _ := keyring.Get(p.app.UniqueID(), "RebellionUsername")
	password, _ := keyring.Get(p.app.UniqueID(), "RebellionPassword")

	return username, password
}

func NewPrefs(a fyne.App) *Prefs {
	return &Prefs{app: a}
}
