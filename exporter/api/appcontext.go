package api

import "fyne.io/fyne/v2/data/binding"

type AppContext struct {
	ShowSettings binding.Bool
}

func NewAppContext() *AppContext {
	return &AppContext{
		ShowSettings: binding.NewBool(),
	}
}
