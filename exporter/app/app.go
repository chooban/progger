package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/exporter/services"
)

type ProggerApp struct {
	AppContext *AppContext
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
		AppContext: NewAppContext(),
		FyneApp:    a,
		RootWindow: w,
		AppService: appServices,
	}
}
