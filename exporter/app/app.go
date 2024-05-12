package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/exporter/prefs"
	"github.com/chooban/progger/exporter/services"
)

type ProggerApp struct {
	State      *State
	Prefs      *prefs.Prefs
	FyneApp    fyne.App
	RootWindow fyne.Window
	AppService *services.AppServices
}

func NewProggerApp() *ProggerApp {
	ctx, _ := context.WithLogger()

	a := app.NewWithID("com.rosshendry.progger.exporter")
	w := a.NewWindow("Progger - Exporter")

	appServices := services.NewAppServices(ctx, a)

	return &ProggerApp{
		State:      NewAppState(appServices),
		FyneApp:    a,
		RootWindow: w,
		AppService: appServices,
	}
}
