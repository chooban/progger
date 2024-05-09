package app

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter/prefs"
	"github.com/chooban/progger/exporter/services"
)

type State struct {
	services      *services.AppServices
	prefs         *prefs.Prefs
	IsDownloading binding.Bool
	IsScanning    binding.Bool
	Stories       binding.UntypedList
}

func (s *State) Dispatch(m interface{}) {
	switch m.(type) {
	case StartScanningMessage:
		_m, _ := m.(StartScanningMessage)
		if err := s.IsScanning.Set(true); err != nil {
			println(err.Error())
		}

		go func() {
			defer func() {
				if err := s.IsScanning.Set(false); err != nil {
					println(err.Error())
				}
			}()

			stories := s.services.Scanner.Scan(_m.Directory)
			_stories := make([]interface{}, len(stories))
			for i, v := range stories {
				_stories[i] = v
			}
			if err := s.Stories.Set(_stories); err != nil {
				println(err.Error())
			}
		}()
	case StartDownloadingMessage:
		s.IsDownloading.Set(true)

		go func() {
			defer func() {
				s.IsDownloading.Set(false)
			}()
			srcDir, _ := s.services.Downloader.BoundSourceDir.Get()
			rUser := s.prefs.RebellionUsername
			rPass := s.prefs.RebellionPassword

			err := s.services.Downloader.Download(srcDir, rUser, rPass)
			s.Dispatch(finishedDownloadingMessage{})

			if err != nil {
				println(err.Error())
			}
		}()
	case finishedDownloadingMessage:
		srcDir, _ := s.services.Downloader.BoundSourceDir.Get()
		s.Dispatch(StartScanningMessage{srcDir})
	}
}

func NewAppState(s *services.AppServices, p *prefs.Prefs) *State {
	c := State{
		services:      s,
		prefs:         p,
		IsDownloading: binding.NewBool(),
		IsScanning:    binding.NewBool(),
		Stories:       binding.NewUntypedList(),
	}

	return &c
}

type StartScanningMessage struct {
	Directory string
}
type StartDownloadingMessage struct{}

type finishedDownloadingMessage struct{}
