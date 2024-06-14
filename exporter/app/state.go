package app

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter/services"
)

type State struct {
	services      *services.AppServices
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
				println("Setting scanning to false")
				if err := s.IsScanning.Set(false); err != nil {
					println(err.Error())
				}
			}()

			stories := s.services.Scanner.Scan(_m.Directory)
			_stories := make([]interface{}, len(stories))
			for i, v := range stories {
				_stories[i] = v
			}
			println("Setting the scanned stories")
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
			srcDir := s.services.Prefs.SourceDirectory()
			rUser, rPass := s.services.Prefs.RebellionDetails()

			if err := s.services.Downloader.Download(srcDir, rUser, rPass); err != nil {
				s.Dispatch(finishedDownloadingMessage{Success: false})
			} else {
				s.Dispatch(finishedDownloadingMessage{Success: true})
			}

		}()
	case finishedDownloadingMessage:
		s.IsScanning.Set(false)
		if m.(finishedDownloadingMessage).Success {
			srcDir := s.services.Prefs.SourceDirectory()
			s.Dispatch(StartScanningMessage{srcDir})
		}
	}
}

func NewAppState(s *services.AppServices) *State {
	c := State{
		services:      s,
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

type finishedDownloadingMessage struct {
	Success bool
}
