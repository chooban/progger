package app

import (
	"fyne.io/fyne/v2/data/binding"
)

type AppContext struct {
	ShowSettings binding.Bool
	dispatcher   Dispatcher
}

func (a *AppContext) Dispatch(m interface{}) {
	a.dispatcher.Dispatch(m)
}

func NewAppContext() *AppContext {
	c := AppContext{
		ShowSettings: binding.NewBool(),
	}
	c.dispatcher = *newDispatcher(&c)

	return &c
}

type Dispatcher struct {
	AppContext *AppContext
}

func (d *Dispatcher) Dispatch(msg interface{}) {
	switch msg.(type) {
	case ShowSettingsMessage:
		d.AppContext.ShowSettings.Set(true)
	case HideSettingsMessage:
		d.AppContext.ShowSettings.Set(false)
	}
}

func newDispatcher(appContext *AppContext) *Dispatcher {
	return &Dispatcher{
		AppContext: appContext,
	}
}

type ShowSettingsMessage struct{}
type HideSettingsMessage struct{}
