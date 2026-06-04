package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2/data/binding"
	"github.com/chooban/progger/database"
	downloadApi "github.com/chooban/progger/download"
	"github.com/chooban/progger/exporter/api"
)

type State struct {
	services       *AppServices
	IsDownloading  binding.Bool
	IsScanning     binding.Bool
	Stories        binding.UntypedList
	AvailableProgs binding.UntypedList
	ToDownload     binding.UntypedList
	SkipTitles     binding.StringList
	KnownTitles    binding.StringList
}

func NewAppState(s *AppServices) *State {
	availableProgs := binding.NewUntypedList()
	appState := State{
		services:       s,
		IsDownloading:  binding.NewBool(),
		IsScanning:     binding.NewBool(),
		Stories:        binding.NewUntypedList(),
		AvailableProgs: availableProgs,
		ToDownload:     binding.NewUntypedList(),
		SkipTitles:     binding.NewStringList(),
		KnownTitles:    binding.NewStringList(),
	}

	if s.DB != nil {
		ctx := context.Background()
		knownRepo := database.NewKnownTitlesRepo(s.DB)
		skipRepo := database.NewSkipTitlesRepo(s.DB)

		if titles, err := knownRepo.List(ctx); err == nil {
			appState.KnownTitles.Set(titles)
		}
		if titles, err := skipRepo.List(ctx); err == nil {
			appState.SkipTitles.Set(titles)
		}
	}

	refreshIssues := func() {
		savedProgs := s.Storage.ReadIssues()
		if len(savedProgs) > 0 {
			convertedProgs := appState.BuildIssueList(savedProgs)
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

func (s *State) GetToDownload() []api.Downloadable {
	items, _ := s.ToDownload.Get()
	result := make([]api.Downloadable, len(items))
	for i, v := range items {
		result[i] = v.(api.Downloadable)
	}
	return result
}

func (s *State) AddToDownload(issue api.Downloadable) {
	items, _ := s.ToDownload.Get()

	for _, v := range items {
		downloadable := v.(api.Downloadable)
		if (&downloadable.Comic).Equals(issue.Comic) {
			return
		}
	}

	s.ToDownload.Append(issue)
}

func (s *State) RemoveFromDownload(issue api.Downloadable) {
	items, _ := s.ToDownload.Get()

	for i, v := range items {
		downloadable := v.(api.Downloadable)
		if (&downloadable.Comic).Equals(issue.Comic) {
			newItems := make([]interface{}, 0, len(items)-1)
			newItems = append(newItems, items[:i]...)
			newItems = append(newItems, items[i+1:]...)
			s.ToDownload.Set(newItems)
			return
		}
	}
}

func (s *State) ClearToDownload() {
	s.ToDownload.Set(make([]interface{}, 0))
}

func (s *State) RefreshProgList() {
	availableProgs, _ := s.AvailableProgs.Get()

	savedProgs := make([]api.Downloadable, len(availableProgs))
	for i, v := range availableProgs {
		savedProgs[i] = v.(api.Downloadable)
	}

	convertedProgs := s.BuildIssueList(savedProgs)

	err := s.AvailableProgs.Set(convertedProgs)
	if err != nil {
		println(err.Error())
	}
}

func (s *State) BuildIssueList(issues []api.Downloadable) []interface{} {
	if len(issues) == 0 {
		return make([]interface{}, 0)
	}

	progSourceDir := s.services.Prefs.ProgSourceDirectory()
	megSourceDir := s.services.Prefs.MegSourceDirectory()

	sort.Slice(issues, func(a, b int) bool {
		return issues[a].Comic.IssueNumber > issues[b].Comic.IssueNumber
	})

	untypedIssues := make([]interface{}, len(issues))
	for i, v := range issues {
		targetDir := progSourceDir
		if v.Comic.Publication == "Megazine" {
			targetDir = megSourceDir
		}

		filename := v.Comic.Filename(downloadApi.Pdf)
		if _, err := os.Stat(filepath.Join(targetDir, filename)); err == nil {
			v.Downloaded = true
		}
		untypedIssues[i] = v
	}

	return untypedIssues
}
