package app

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter/prefs"
	"github.com/chooban/progger/exporter/services"
)

type State struct {
	ShowSettings binding.Bool
	Services     *services.AppServices
	prefs        *prefs.Prefs
}

func (a *State) Dispatch(m interface{}) {
	switch m.(type) {
	case ShowSettingsMessage:
		a.ShowSettings.Set(true)
	case HideSettingsMessage:
		a.ShowSettings.Set(false)
	case StartScanningMessage:
		_m, _ := m.(StartScanningMessage)
		go func() {
			a.Services.Scanner.Scan(_m.Directory)
		}()
	case StartDownloadingMessage:
		go func() {
			srcDir, _ := a.Services.Downloader.BoundSourceDir.Get()
			rUser := a.prefs.RebellionUsername
			rPass := a.prefs.RebellionPassword

			if err := a.Services.Downloader.Download(srcDir, rUser, rPass); err == nil {
				a.Dispatch(finishedDownloadingMessage{})
			} else {
				println(err.Error())
			}
		}()
	case finishedDownloadingMessage:
		srcDir, _ := a.Services.Downloader.BoundSourceDir.Get()
		a.Dispatch(StartScanningMessage{srcDir})
	}
}

func NewAppState(s *services.AppServices, p *prefs.Prefs) *State {
	c := State{
		ShowSettings: binding.NewBool(),
		Services:     s,
		prefs:        p,
	}

	return &c
}

type ShowSettingsMessage struct{}
type HideSettingsMessage struct{}
type StartScanningMessage struct {
	Directory string
}
type StartDownloadingMessage struct{}

type finishedDownloadingMessage struct{}
