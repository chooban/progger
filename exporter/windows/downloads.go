package windows

import (
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
)

func newDownloadsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	progress := newDownloadProgress()
	dButton := downloadButton(a, "Download Prog List")
	progListContainer := newProgListContainer(a.State, a)

	mainPanel := container.New(
		layout.NewStackLayout(),
		progress,
		dButton,
		progListContainer,
	)

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := a.State.IsDownloading.Get()
		if isDownloading {
			showHide(mainPanel, progress)
		} else {
			downloads, _ := a.State.AvailableProgs.Get()
			if len(downloads) > 0 {
				showHide(mainPanel, progListContainer)
			} else {
				showHide(mainPanel, dButton)
			}
		}
	}))

	c := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		mainPanel,
	)

	return c
}

func newProgListContainer(s *app.State, proggerApp *app.ProggerApp) *fyne.Container {
	progs := s.AvailableProgs
	refreshDownloadListButton := widget.NewButton("Refresh Downloads List", func() {
		startFetchIssuesList(proggerApp)
	})
	downloadAllButton := widget.NewButton("Download Selected", func() {
		startDownloadSelected(proggerApp)
	})

	nothingToDownload := widget.NewLabel("No new progs to download")

	onChecked := func(issue api.Downloadable, shouldDownload bool) {
		// This fires when we programmatically change the state in the list as well. Since the
		// list items are reused, it's always firing. We need to compensate for that.

		// If we've already downloaded the issue, then don't do anything. There's no functionality
		// for deleting or re-downloading
		if !issue.Downloaded {
			if shouldDownload {
				s.AddToDownload(issue)
			} else {
				s.RemoveFromDownload(issue)
			}
		}
	}

	isMarked := func(issue api.Downloadable) bool {
		items, _ := s.ToDownload.Get()
		for _, v := range items {
			downloadable := v.(api.Downloadable)
			if (&downloadable.Comic).Equals(issue.Comic) {
				return true
			}
		}
		return false
	}

	progListWidget := newProgList(progs, onChecked, isMarked)
	mainDisplay := container.NewStack(progListWidget, nothingToDownload)

	progs.AddListener(binding.NewDataListener(func() {
		downloadCandidates, _ := progs.Get()
		if len(downloadCandidates) > 0 {
			showHide(mainDisplay, progListWidget)
		} else {
			showHide(mainDisplay, nothingToDownload)
		}
	}))

	buttonsContainer := container.NewHBox(refreshDownloadListButton, downloadAllButton)

	return container.NewBorder(nil, container.NewCenter(buttonsContainer), nil, nil, mainDisplay)
}

type issueToggler func(issue api.Downloadable, shouldDownload bool)
type isMarkedForDownload func(issue api.Downloadable) bool

func newProgList(progs binding.UntypedList, onCheck issueToggler, isMarked isMarkedForDownload) fyne.CanvasObject {
	listOfProgs := widget.NewListWithData(
		progs,
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewCheck("", func(b bool) {}),
				widget.NewLabel(""),
			)
		},
		func(di binding.DataItem, o fyne.CanvasObject) {
			ctr, _ := o.(*fyne.Container)
			diu, _ := di.(binding.Untyped).Get()
			downloadable := diu.(api.Downloadable)

			check := ctr.Objects[1].(*widget.Check)

			check.OnChanged = func(checked bool) {}
			if downloadable.Downloaded {
				check.SetChecked(true)
				check.Disable()
			} else {
				if isMarked(downloadable) {
					check.SetChecked(true)
				} else {
					check.SetChecked(false)
				}
				check.Enable()
			}
			check.OnChanged = func(checked bool) {
				onCheck(downloadable, checked)
			}

			if reflect.TypeOf(ctr.Objects[0]).String() == "*widget.Label" {
				label := ctr.Objects[0].(*widget.Label)
				label.SetText(downloadable.Comic.String())
			} else {
				label := widget.NewLabel(downloadable.Comic.String())
				ctr.Objects[0] = label
			}
		},
	)

	c := container.NewBorder(
		nil, nil, nil, nil, listOfProgs,
	)

	return c
}

func newDownloadProgress() *fyne.Container {
	activity := widget.NewActivity()
	activity.Start()

	mainPanel := container.New(
		layout.NewVBoxLayout(),
		activity,
		widget.NewLabel("Downloading..."),
	)

	return container.NewCenter(mainPanel)
}

func downloadButton(a *app.ProggerApp, label string) fyne.CanvasObject {
	downloadButton := widget.NewButton(label, func() {
		startFetchIssuesList(a)
	})

	return container.NewCenter(downloadButton)
}

func startFetchIssuesList(a *app.ProggerApp) {
	rUser, rPass := a.Services.Prefs.RebellionDetails()

	// Create the operation
	op := app.NewDownloadListOperation()

	ctx, cancel, _ := app.WithLogger()
	op.SetCancel(cancel)

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		issues, err := a.Services.Downloader.FetchIssuesList(ctx, rUser, rPass)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Convert to []interface{} for binding
		untypedProgs := make([]interface{}, len(issues))
		for i, v := range issues {
			untypedProgs[i] = v
		}

		if err := op.AvailableProgs.Set(untypedProgs); err != nil {
			_ = op.Error.Set(err.Error())
			return
		}
	}()

	// Bind the operation state to the app state
	op.IsRunning.AddListener(binding.NewDataListener(func() {
		isRunning, _ := op.IsRunning.Get()
		a.State.IsDownloading.Set(isRunning)
	}))

	op.AvailableProgs.AddListener(binding.NewDataListener(func() {
		progs, _ := op.AvailableProgs.Get()
		// Convert from []interface{} to []api.Downloadable
		downloadables := make([]api.Downloadable, len(progs))
		for i, v := range progs {
			downloadables[i] = v.(api.Downloadable)
		}
		// Transform with download status checking and sorting
		transformed := a.State.BuildIssueList(downloadables)
		a.State.AvailableProgs.Set(transformed)
	}))
}

func startDownloadSelected(a *app.ProggerApp) {
	rUser, rPass := a.Services.Prefs.RebellionDetails()
	progSourceDir := a.Services.Prefs.ProgSourceDirectory()
	megSourceDir := a.Services.Prefs.MegSourceDirectory()

	toDownload := a.State.GetToDownload()
	if len(toDownload) == 0 {
		return
	}

	// Create the operation
	op := app.NewDownloadOperation()

	ctx, cancel, _ := app.WithLogger()
	op.SetCancel(cancel)

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		err := a.Services.Downloader.DownloadIssues(ctx, toDownload, progSourceDir, megSourceDir, rUser, rPass)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}
	}()

	// Bind the operation state to the app state
	op.IsRunning.AddListener(binding.NewDataListener(func() {
		isRunning, _ := op.IsRunning.Get()
		a.State.IsDownloading.Set(isRunning)

		// When download completes, clear the download list and refresh
		if !isRunning {
			a.State.ClearToDownload()
			a.State.RefreshProgList()

			// Also trigger a scan
			triggerScanAfterDownload(a)
		}
	}))
}

func triggerScanAfterDownload(a *app.ProggerApp) {
	dirsToScan := []string{a.Services.Prefs.ProgSourceDirectory(), a.Services.Prefs.MegSourceDirectory()}
	knownTitles := a.Services.Storage.ReadKnownTitles()
	skipTitles := a.Services.Storage.ReadSkipTitles()

	// Create the operation
	op := app.NewScanOperation()

	ctx, cancel, _ := app.WithLogger()
	op.SetCancel(cancel)

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		foundStories, err := a.Services.Scanner.Scan(ctx, dirsToScan, knownTitles, skipTitles)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Convert to untyped for binding
		untypedStories := make([]interface{}, len(foundStories))
		storiesToStore := make([]api.Story, len(foundStories))
		for i, v := range foundStories {
			untypedStories[i] = v
			storiesToStore[i] = *v
		}

		if err := op.Stories.Set(untypedStories); err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Store the stories
		if err := a.Services.Storage.StoreStories(storiesToStore); err != nil {
			_ = op.Error.Set("Failed to save stories: " + err.Error())
		}
	}()

	// Bind the operation state to the app state
	op.IsRunning.AddListener(binding.NewDataListener(func() {
		isRunning, _ := op.IsRunning.Get()
		a.State.IsScanning.Set(isRunning)
	}))

	op.Stories.AddListener(binding.NewDataListener(func() {
		stories, _ := op.Stories.Get()
		a.State.Stories.Set(stories)
	}))
}
