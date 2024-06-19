package app

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/exporter/services"
)

type State struct {
	services       *services.AppServices
	IsDownloading  binding.Bool
	IsScanning     binding.Bool
	Stories        binding.UntypedList
	AvailableProgs binding.UntypedList
}

func (s *State) startScanningHandler(m StartScanningMessage) {
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

		stories := s.services.Scanner.Scan(m.Directory)
		_stories := make([]interface{}, len(stories))
		for i, v := range stories {
			_stories[i] = v
		}
		println("Setting the scanned stories")
		if err := s.Stories.Set(_stories); err != nil {
			println(err.Error())
		}
	}()
}

func (s *State) startDownloadingHandler(_m StartDownloadingMessage) {
	s.IsDownloading.Set(true)

	go func() {
		defer func() {
			s.IsDownloading.Set(false)
		}()
		srcDir := s.services.Prefs.SourceDirectory()
		rUser, rPass := s.services.Prefs.RebellionDetails()

		ctx, _ := context.WithLogger()
		if err := s.services.Downloader.DownloadAllProgs(ctx, srcDir, rUser, rPass); err != nil {
			s.Dispatch(finishedDownloadingMessage{Success: false})
		} else {
			s.Dispatch(finishedDownloadingMessage{Success: true})
		}

	}()
}

func (s *State) downloadProgListHandler(m StartDownloadingProgListMessage) {
	// IsDownloading is pretty much a synonym for "is interacting with Rebellion account"
	s.IsDownloading.Set(true)

	storedProgs := s.services.Storage.ReadProgs()
	if m.Refresh == false && len(storedProgs) > 0 {
		println("Found progs")
		progs := make([]interface{}, len(storedProgs))
		for i, v := range storedProgs {
			progs[i] = v
		}
		err := s.AvailableProgs.Set(progs)
		if err != nil {
			println(err.Error())
		}
		s.Dispatch(finishedDownloadingMessage{Success: true})
		s.IsDownloading.Set(false)
		return
	}

	go func() {
		defer func() {
			s.IsDownloading.Set(false)
		}()
		rUser, rPass := s.services.Prefs.RebellionDetails()

		ctx, _ := context.WithLogger()
		if list, err := s.services.Downloader.ProgList(ctx, rUser, rPass); err != nil {
			s.Dispatch(finishedDownloadingMessage{Success: false})
		} else {
			progs := make([]interface{}, len(list))
			for i, v := range list {
				progs[i] = v
			}
			s.AvailableProgs.Set(progs)
			err := s.services.Storage.SaveProgs(list)
			if err != nil {
				println(err.Error())
			}
			s.Dispatch(finishedDownloadingMessage{Success: true})
		}
	}()
}

func (s *State) Dispatch(m interface{}) {
	switch m.(type) {
	case StartScanningMessage:
		s.startScanningHandler(m.(StartScanningMessage))
	case StartDownloadingMessage:
		s.startDownloadingHandler(m.(StartDownloadingMessage))
	case StartDownloadingProgListMessage:
		s.downloadProgListHandler(m.(StartDownloadingProgListMessage))
	case finishedDownloadingMessage:
		if m.(finishedDownloadingMessage).Success {
			srcDir := s.services.Prefs.SourceDirectory()
			s.Dispatch(StartScanningMessage{srcDir})
		}
	}
}

func NewAppState(s *services.AppServices) *State {
	c := State{
		services:       s,
		IsDownloading:  binding.NewBool(),
		IsScanning:     binding.NewBool(),
		Stories:        binding.NewUntypedList(),
		AvailableProgs: binding.NewUntypedList(),
	}

	return &c
}

type StartScanningMessage struct {
	Directory string
}

// StartDownloadingMessage requests that all available progs be downloaded
type StartDownloadingMessage struct{}

// StartDownloadingProgListMessage requests downloading a list of available progs from the Rebellion account
type StartDownloadingProgListMessage struct {
	Refresh bool
}

type finishedDownloadingMessage struct {
	Success bool
}
