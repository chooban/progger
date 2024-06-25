package app

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/exporter/services"
	"sort"
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

	go func() {
		defer func() {
			s.IsDownloading.Set(false)
		}()
		rUser, rPass := s.services.Prefs.RebellionDetails()

		ctx, _ := context.WithLogger()
		if list, err := s.services.Downloader.ProgList(ctx, rUser, rPass); err != nil {
			s.Dispatch(finishedDownloadingMessage{Success: false})
		} else {
			downloadableList := make([]api.Downloadable, len(list))
			for i, v := range list {
				p := api.Downloadable{
					Comic:      v,
					Downloaded: false,
				}
				downloadableList[i] = p
			}
			progs := s.buildProgList(downloadableList)
			s.AvailableProgs.Set(progs)
			err := s.services.Storage.SaveProgs(downloadableList)
			if err != nil {
				println(err.Error())
			}
			s.Dispatch(finishedDownloadingMessage{Success: true})
		}
	}()
}

func (s *State) buildProgList(progs []api.Downloadable) []interface{} {
	if len(progs) == 0 {
		return make([]interface{}, 0)
	}
	sort.Slice(progs, func(a, b int) bool {
		return progs[a].Comic.IssueNumber > progs[b].Comic.IssueNumber
	})
	println(fmt.Sprintf("Checking %s for progs", s.services.Prefs.SourceDirectory()))
	untypedProgs := make([]interface{}, len(progs))
	for i, v := range progs {
		untypedProgs[i] = v
	}

	return untypedProgs
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
	availableProgs := binding.NewUntypedList()
	c := State{
		services:       s,
		IsDownloading:  binding.NewBool(),
		IsScanning:     binding.NewBool(),
		Stories:        binding.NewUntypedList(),
		AvailableProgs: availableProgs,
	}

	refreshProgs := func() {
		println("Refreshing progs")
		savedProgs := s.Storage.ReadProgs()
		if len(savedProgs) > 0 {
			convertedProgs := c.buildProgList(savedProgs)
			if len(convertedProgs) > 0 {
				println(fmt.Sprintf("Found %d progs", len(convertedProgs)))
				if err := availableProgs.Set(convertedProgs); err == nil {
					// Do nothing
				}
			}
		}
	}
	c.services.Prefs.BoundSourceDir.AddListener(binding.NewDataListener(refreshProgs))

	refreshProgs()

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
