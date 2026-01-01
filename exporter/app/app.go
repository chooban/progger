package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type ProggerApp struct {
	State      *State
	FyneApp    fyne.App
	RootWindow fyne.Window
	Services   *AppServices
}

func NewProggerApp() *ProggerApp {
	a := app.NewWithID("com.rosshendry.progger.exporter")
	w := a.NewWindow("Progger - Exporter")

	appServices := NewAppServices(a)

	return &ProggerApp{
		State:      NewAppState(appServices),
		FyneApp:    a,
		RootWindow: w,
		Services:   appServices,
	}
}
