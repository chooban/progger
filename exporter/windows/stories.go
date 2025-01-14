package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/context"
	"slices"
	"strings"
)

func showHide(container *fyne.Container, toShow fyne.CanvasObject) {
	for i := 0; i < len(container.Objects); i++ {
		container.Objects[i].Hide()
	}
	toShow.Show()
}

func newStoriesCanvas(a *app.ProggerApp) fyne.CanvasObject {
	storiesPanel := storiesContainer(a)
	scannerProgress := newScannerProgressContainer()
	downloadProgress := newDownloadProgressContainer()

	centralLayout := container.New(
		layout.NewStackLayout(),
		storiesPanel,
		scannerProgress,
		downloadProgress,
	)

	a.State.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := a.State.IsScanning.Get(); isScanning == true {
			showHide(centralLayout, scannerProgress)
		} else {
			showHide(centralLayout, storiesPanel)
		}
	}))

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		if isDownloading, _ := a.State.IsDownloading.Get(); isDownloading == true {
			println("Showing the download progress")
			showHide(centralLayout, downloadProgress)
		} else {
			showHide(centralLayout, storiesPanel)
		}
	}))

	storiesLayout := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		centralLayout,
	)

	if isDownloading, _ := a.State.IsDownloading.Get(); isDownloading == true {
		println("Showing the download progress")
		showHide(centralLayout, downloadProgress)
	} else {
		println("Showing the stories panel")
		showHide(centralLayout, storiesPanel)
	}

	return storiesLayout
}

func newScannerProgressContainer() *fyne.Container {
	barContainer := container.NewVBox(
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Scanning..."),
	)
	centeredBar := container.NewCenter(
		barContainer,
	)

	return centeredBar
}

func newStoryListWidget(boundStories binding.UntypedList) *fyne.Container {
	filterValue := binding.NewString()
	filteredList := binding.NewUntypedList()

	storyList := widget.NewListWithData(
		filteredList,
		// Component structure of the row
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewCheck("", func(b bool) {}),
				// takes the rest of the space
				widget.NewLabel(""),
			)
		},
		func(di binding.DataItem, o fyne.CanvasObject) {
			ctr, _ := o.(*fyne.Container)
			// ideally we should check `ok` for each one of those casting
			// but we know that they are those types for sure
			label := ctr.Objects[0].(*widget.Label)
			check := ctr.Objects[1].(*widget.Check)
			diu, _ := di.(binding.Untyped).Get()
			story := diu.(*api.Story)

			b := binding.BindBool(&story.ToExport)
			label.SetText(fmt.Sprintf("%s (%s)", story.Display(), story.IssueSummary()))
			check.Bind(b)
		},
	)

	listRefresh := func() {
		stories, _ := boundStories.Get()
		toDisplay := make([]interface{}, 0, len(stories))

		f, _ := filterValue.Get()
		if strings.TrimSpace(f) == "" {
			toDisplay = slices.Clone(stories)
		} else {
			_f := strings.Split(strings.ToLower(f), " ")
			for _, v := range stories {
				if ContainsAll(strings.ToLower(v.(*api.Story).Display()), _f) {
					toDisplay = append(toDisplay, v)
				}
			}
		}

		if err := filteredList.Set(toDisplay); err != nil {
			println(err.Error())
		}
		storyList.Refresh()
	}

	boundStories.AddListener(binding.NewDataListener(listRefresh))
	filterValue.AddListener(binding.NewDataListener(listRefresh))

	filter := widget.NewEntryWithData(filterValue)
	filter.SetPlaceHolder("Filter list")

	c := container.NewBorder(
		nil,
		filter,
		nil,
		nil,
		storyList,
	)

	return c
}

func noStoriesContainer(a *app.ProggerApp) fyne.CanvasObject {
	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan := a.AppService.Prefs.SourceDirectory()
		a.State.Dispatch(app.StartScanningMessage{Directory: dirToScan})
	})
	content := container.NewVBox(widget.NewLabelWithData(a.AppService.Prefs.BoundSourceDir), scanButton)

	contentWrapper := container.NewCenter(
		content,
	)

	return contentWrapper
}

func storiesContainer(a *app.ProggerApp) fyne.CanvasObject {
	listContainer := container.NewBorder(
		nil, storiesButtonsContainer(a), nil, nil,
		newStoryListWidget(a.State.Stories),
	)
	noListContainer := noStoriesContainer(a)

	c := container.New(layout.NewStackLayout(), listContainer, noListContainer)

	boundStories := a.State.Stories
	boundStories.AddListener(binding.NewDataListener(func() {
		stories, _ := boundStories.Get()
		if len(stories) == 0 {
			showHide(c, noListContainer)
		} else {
			showHide(c, listContainer)
		}
	}))

	return c
}

func newDownloadProgressContainer() *fyne.Container {
	barContainer := container.NewVBox(
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)
	centeredBar := container.NewCenter(
		barContainer,
	)

	return centeredBar
}

func storiesButtonsContainer(a *app.ProggerApp) fyne.CanvasObject {
	exportButton := exportButton(a)

	scanButton := widget.NewButton("Force Rescan", func() {
		dirToScan := a.AppService.Prefs.SourceDirectory()
		a.State.Dispatch(app.StartScanningMessage{Directory: dirToScan})
	})

	return container.NewVBox(
		exportButton,
		scanButton,
	)
}

func exportButton(a *app.ProggerApp) *widget.Button {
	exporter := a.AppService.Exporter
	prefsService := a.AppService.Prefs

	exportButton := widget.NewButton("Export Story", func() {
		stories, err := a.State.Stories.Get()
		if err != nil {
			dialog.ShowError(err, a.RootWindow)
		}
		toExport := make([]*api.Story, 0)
		for _, v := range stories {
			story := v.(*api.Story)
			if story.ToExport {
				toExport = append(toExport, story)
			}
		}
		if len(toExport) == 0 {
			dialog.ShowInformation("Export", "No stories selected", a.RootWindow)
		} else {
			filename := binding.NewString()
			filename.Set(toExport[0].Display() + ".pdf")
			fnameEntry := widget.NewEntryWithData(filename)

			onClose := func(b bool) {
				if b {
					fname, _ := filename.Get()
					ctx, _ := context.WithLogger()
					if err := exporter.Export(ctx, toExport, prefsService.SourceDirectory(), prefsService.ExportDirectory(), fname); err != nil {
						dialog.ShowError(err, a.RootWindow)
					} else {
						dialog.ShowInformation("Export", "File successfully exported", a.RootWindow)
					}
				}
			}

			formDialog := dialog.NewForm(
				"Export",
				"Export",
				"Cancel",
				[]*widget.FormItem{
					{Text: "Filename", Widget: fnameEntry},
				},
				onClose,
				a.RootWindow,
			)
			formDialog.Show()
			formDialog.Resize(fyne.NewSize(500, 100))
		}
	})

	return exportButton
}

func ContainsAll(s string, t []string) bool {
	if len(t) == 0 {
		return true
	}
	for _, v := range t {
		if !strings.Contains(s, v) {
			return false
		}
	}
	return true
}
