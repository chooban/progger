package app

import (
	"os"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2/data/binding"
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

	// Look for stored known titles
	storedKnownTitles := s.Storage.ReadKnownTitles()
	appState.KnownTitles.Set(storedKnownTitles)

	// Look for stored skip titles
	storedSkipTitles := s.Storage.ReadSkipTitles()
	appState.SkipTitles.Set(storedSkipTitles)

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

// GetToDownload returns the current list of items to download
func (s *State) GetToDownload() []api.Downloadable {
	items, _ := s.ToDownload.Get()
	result := make([]api.Downloadable, len(items))
	for i, v := range items {
		result[i] = v.(api.Downloadable)
	}
	return result
}

// AddToDownload adds an item to the download list if not already present
func (s *State) AddToDownload(issue api.Downloadable) {
	items, _ := s.ToDownload.Get()

	// Check if already in list
	for _, v := range items {
		downloadable := v.(api.Downloadable)
		if (&downloadable.Comic).Equals(issue.Comic) {
			return
		}
	}

	s.ToDownload.Append(issue)
}

// RemoveFromDownload removes an item from the download list
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

// ClearToDownload clears the download list
func (s *State) ClearToDownload() {
	s.ToDownload.Set(make([]interface{}, 0))
}

// RefreshProgList refreshes the available progs list to mark downloaded items
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

// BuildIssueList converts issues to untyped list, checks which are already downloaded, and sorts
func (s *State) BuildIssueList(issues []api.Downloadable) []interface{} {
	if len(issues) == 0 {
		return make([]interface{}, 0)
	}

	progSourceDir := s.services.Prefs.ProgSourceDirectory()
	megSourceDir := s.services.Prefs.MegSourceDirectory()

	// Sort by issue number descending
	sort.Slice(issues, func(a, b int) bool {
		return issues[a].Comic.IssueNumber > issues[b].Comic.IssueNumber
	})

	untypedIssues := make([]interface{}, len(issues))
	for i, v := range issues {
		targetDir := progSourceDir
		if v.Comic.Publication == "Megazine" {
			targetDir = megSourceDir
		}

		// Check if already downloaded
		filename := v.Comic.Filename(downloadApi.Pdf)
		if _, err := os.Stat(filepath.Join(targetDir, filename)); err == nil {
			v.Downloaded = true
		}
		untypedIssues[i] = v
	}

	return untypedIssues
}
