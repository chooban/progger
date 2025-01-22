package app

import (
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	downloadApi "github.com/chooban/progger/download/api"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/context"
	"github.com/chooban/progger/exporter/services"
	"os"
	"path/filepath"
	"slices"
	"sort"
)

type State struct {
	services       *services.AppServices
	IsDownloading  binding.Bool
	IsScanning     binding.Bool
	Stories        binding.UntypedList
	AvailableProgs binding.UntypedList
	ToDownload     []api.Downloadable
}

func (s *State) startScanningHandler(m StartScanningMessage) {
	if err := s.IsScanning.Set(true); err != nil {
		println(err.Error())
	}

	go func() {
		defer func() {
			if err := s.IsScanning.Set(false); err != nil {
				println(err.Error())
			}
		}()

		dirsToScan := []string{s.services.Prefs.ProgSourceDirectory(), s.services.Prefs.MegSourceDirectory()}
		foundStories := s.services.Scanner.Scan(dirsToScan)
		untypedStories := make([]interface{}, len(foundStories))
		storiesToStore := make([]api.Story, len(foundStories))
		for i, v := range foundStories {
			untypedStories[i] = v
			storiesToStore[i] = *v
		}
		if err := s.Stories.Set(untypedStories); err != nil {
			println(err.Error())
		}
		s.services.Storage.StoreStories(storiesToStore)
	}()
}

func (s *State) startDownloadingHandler(_m StartDownloadingMessage) {
	s.IsDownloading.Set(true)

	go func() {
		defer func() {
			s.IsDownloading.Set(false)
		}()
		srcDir := s.services.Prefs.ProgSourceDirectory()
		rUser, rPass := s.services.Prefs.RebellionDetails()

		ctx, _ := context.WithLogger()
		if err := s.services.Downloader.DownloadAllIssues(ctx, srcDir, rUser, rPass); err != nil {
			s.Dispatch(finishedDownloadingMessage{Success: false})
		} else {
			s.Dispatch(finishedDownloadingMessage{Success: true})
		}

	}()
}

// refreshProgList loops through the list of available issues and marks those we have as downloaded
func (s *State) refreshProgList() {
	availableProgs, _ := s.AvailableProgs.Get()

	var issue api.Downloadable
	for i, v := range availableProgs {
		issue = v.(api.Downloadable)
		if _, err := os.Stat(filepath.Join(s.services.Prefs.ProgSourceDirectory(), issue.Comic.Filename(downloadApi.Pdf))); err == nil {
			issue.Downloaded = true

			availableProgs[i] = issue
		}
	}

	err := s.AvailableProgs.Set(availableProgs)
	if err != nil {
		println(err.Error())
	}
}

func (s *State) downloadSelectedProgs(m DownloadSelectedMessage) {
	if err := s.IsDownloading.Set(true); err != nil {
		println(err.Error())
	}

	go func() {
		defer func() {
			if err := s.IsDownloading.Set(false); err != nil {
				println(err.Error())
			}
			s.ToDownload = make([]api.Downloadable, 0)
			s.refreshProgList()
		}()
		rUser, rPass := s.services.Prefs.RebellionDetails()

		ctx, _ := context.WithLogger()

		for _, v := range s.ToDownload {
			targetDir := s.services.Prefs.ProgSourceDirectory()
			if v.Comic.Publication == "Megazine" {
				targetDir = s.services.Prefs.MegSourceDirectory()
			}
			if err := s.services.Downloader.DownloadIssue(ctx, v.Comic, targetDir, rUser, rPass); err != nil {
				println(err.Error())
				return
			} else {
				v.Downloaded = true
			}
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
		if list, err := s.services.Downloader.GetIssuesList(ctx, rUser, rPass); err != nil {
			s.Dispatch(finishedDownloadingMessage{Success: false})
		} else {
			downloadableList := make([]api.Downloadable, 0, len(list))
			for _, v := range list {
				p := api.Downloadable{
					Comic:      v,
					Downloaded: false,
				}
				downloadableList = append(downloadableList, p)
			}
			progs := s.buildIssueList(downloadableList)
			s.AvailableProgs.Set(progs)
			err := s.services.Storage.SaveIssues(downloadableList)
			if err != nil {
				println(err.Error())
			}
			s.Dispatch(finishedDownloadingMessage{Success: true})
		}
	}()
}

func (s *State) buildIssueList(issues []api.Downloadable) []interface{} {
	if len(issues) == 0 {
		return make([]interface{}, 0)
	}
	sort.Slice(issues, func(a, b int) bool {
		return issues[a].Comic.IssueNumber > issues[b].Comic.IssueNumber
	})
	println(fmt.Sprintf("Checking %s for issues", s.services.Prefs.ProgSourceDirectory()))
	untypedIssues := make([]interface{}, len(issues))
	for i, v := range issues {
		targetDir := s.services.Prefs.ProgSourceDirectory()
		if v.Comic.Publication == "Megazine" {
			targetDir = s.services.Prefs.MegSourceDirectory()
		}
		if _, err := os.Stat(filepath.Join(targetDir, v.Comic.Filename(downloadApi.Pdf))); err == nil {
			v.Downloaded = true
		}
		untypedIssues[i] = v
	}

	return untypedIssues
}

func (s *State) Dispatch(m interface{}) {
	switch m.(type) {
	case StartScanningMessage:
		s.startScanningHandler(m.(StartScanningMessage))
	case StartDownloadingMessage:
		s.startDownloadingHandler(m.(StartDownloadingMessage))
	case StartDownloadingProgListMessage:
		s.downloadProgListHandler(m.(StartDownloadingProgListMessage))
	case DownloadSelectedMessage:
		s.downloadSelectedProgs(m.(DownloadSelectedMessage))
	case AddToDownloadsMessage:
		_m := m.(AddToDownloadsMessage)
		idx := slices.IndexFunc(s.ToDownload, func(downloadable api.Downloadable) bool {
			return downloadable.Comic.Equals(_m.Issue.Comic)
		})
		if idx < 0 {
			s.ToDownload = append(s.ToDownload, _m.Issue)
		}

		println(fmt.Sprintf("To download is: %+v", s.ToDownload))
	case RemoveFromDownloadsMessage:
		_m := m.(RemoveFromDownloadsMessage)
		idx := slices.IndexFunc(s.ToDownload, func(downloadable api.Downloadable) bool {
			return downloadable.Comic.Equals(_m.Issue.Comic)
		})
		if idx >= 0 {
			s.ToDownload[idx] = s.ToDownload[len(s.ToDownload)-1]
			s.ToDownload = s.ToDownload[:len(s.ToDownload)-1]
		}
	case finishedDownloadingMessage:
		if m.(finishedDownloadingMessage).Success {
			s.Dispatch(StartScanningMessage{})
		}
	}
}

func NewAppState(s *services.AppServices) *State {
	availableProgs := binding.NewUntypedList()
	appState := State{
		services:       s,
		IsDownloading:  binding.NewBool(),
		IsScanning:     binding.NewBool(),
		Stories:        binding.NewUntypedList(),
		AvailableProgs: availableProgs,
		ToDownload:     make([]api.Downloadable, 0),
	}

	storedStories := s.Storage.ReadStories()
	untypedStories := make([]interface{}, len(storedStories))
	for i, v := range storedStories {
		untypedStories[i] = &v
	}
	appState.Stories.Set(untypedStories)

	refreshIssues := func() {
		savedProgs := s.Storage.ReadIssues()
		if len(savedProgs) > 0 {
			convertedProgs := appState.buildIssueList(savedProgs)
			if len(convertedProgs) > 0 {
				if err := availableProgs.Set(convertedProgs); err != nil {
					println(err.Error())
				}
			}
		}
	}

	appState.services.Prefs.ProgSourceDir.AddListener(binding.NewDataListener(refreshIssues))
	appState.services.Prefs.MegazineSourceDir.AddListener(binding.NewDataListener(refreshIssues))

	refreshIssues()

	return &appState
}

type StartScanningMessage struct{}

// StartDownloadingMessage requests that all available progs be downloaded
type StartDownloadingMessage struct{}

// DownloadSelectedMessage requests that the selected issues be downloaded
type DownloadSelectedMessage struct{}

// StartDownloadingProgListMessage requests downloading a list of available progs from the Rebellion account
type StartDownloadingProgListMessage struct {
	Refresh bool
}

type finishedDownloadingMessage struct {
	Success bool
}

type AddToDownloadsMessage struct {
	Issue api.Downloadable
}

type RemoveFromDownloadsMessage struct {
	Issue api.Downloadable
}
