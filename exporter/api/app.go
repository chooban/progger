package api

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type ProggerApp struct {
	AppContext *AppContext
	FyneApp    fyne.App
	RootWindow fyne.Window
}

func NewProggerApp() *ProggerApp {
	a := app.NewWithID("com.rosshendry.progger.exporter")
	w := a.NewWindow("Progger - Exporter")

	return &ProggerApp{
		AppContext: NewAppContext(),
		FyneApp:    a,
		RootWindow: w,
	}
}
