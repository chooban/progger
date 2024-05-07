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

	username := a.Preferences().String("RebellionUsername")
	password := a.Preferences().String("RebellionPassword")

	p := &prefs.Prefs{
		RebellionUsername: username,
		RebellionPassword: password,
	}
	appServices := services.NewAppServices(ctx, a, p)

	return &ProggerApp{
		State:      NewAppState(appServices, p),
		FyneApp:    a,
		RootWindow: w,
		AppService: appServices,
		Prefs:      p,
	}
}
