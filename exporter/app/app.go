package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/chooban/progger/exporter/services"
)

type ProggerApp struct {
	State      *State
	FyneApp    fyne.App
	RootWindow fyne.Window
	AppService *services.AppServices
}

func NewProggerApp() *ProggerApp {
	a := app.NewWithID("com.rosshendry.progger.exporter")
	w := a.NewWindow("Progger - Exporter")

	appServices := services.NewAppServices(a)

	return &ProggerApp{
		State:      NewAppState(appServices),
		FyneApp:    a,
		RootWindow: w,
		AppService: appServices,
	}
}
