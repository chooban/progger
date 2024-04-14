package exporter

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/prefs"
	"github.com/chooban/progger/exporter/services"
	"slices"
	"strings"
)

func MainWindow(a fyne.App, w fyne.Window) fyne.CanvasObject {
	ctx, _ := WithLogger()
	appServices := services.NewAppServices(ctx, a)
	boundSource := prefs.BoundSourceDir(a)

	scannerButtonsPanel := buttonsContainer(w, boundSource, appServices)
	displayPanel := displayContainer(boundSource, appServices)

	return container.NewBorder(
		container.NewCenter(
			widget.NewLabel("Borag Thungg!"),
		),
		scannerButtonsPanel,
		nil,
		nil,
		displayPanel,
	)
}

func newScannerContainer() *fyne.Container {
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
			l := ctr.Objects[0].(*widget.Label)
			c := ctr.Objects[1].(*widget.Check)
			diu, _ := di.(binding.Untyped).Get()
			story := diu.(*api.Story)

			b := binding.BindBool(&story.ToExport)
			l.SetText(fmt.Sprintf("%s (%s)", story.Display(), story.IssueSummary()))
			c.Bind(b)
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

func displayContainer(boundSource binding.String, appServices *services.AppServices) fyne.CanvasObject {
	scanner := appServices.Scanner
	downloader := appServices.Downloader

	scannerProgressContainer := newScannerContainer()
	sourceDirectoryLabel := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithData(boundSource),
		),
	)

	listContainer := container.NewStack(newStoryListWidget(scanner.BoundStories))
	listContainer.Hide()

	layout := container.NewStack(
		sourceDirectoryLabel,
		scannerProgressContainer,
		listContainer,
	)

	downloader.IsDownloading.AddListener(binding.NewDataListener(func() {
		if isDownloading, _ := downloader.IsDownloading.Get(); isDownloading == true {
			sourceDirectoryLabel.Hide()
			scannerProgressContainer.Show()
		} else {
			sourceDirectoryLabel.Show()
			scannerProgressContainer.Show()
		}
	}))

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := scanner.IsScanning.Get(); isScanning == true {
			sourceDirectoryLabel.Hide()
			scannerProgressContainer.Show()
		} else {
			stories, _ := scanner.BoundStories.Get()
			if len(stories) == 0 {
				sourceDirectoryLabel.Show()
				scannerProgressContainer.Hide()
			} else {
				sourceDirectoryLabel.Hide()
				scannerProgressContainer.Hide()
				listContainer.Show()
			}
		}
	}))

	return layout
}

func buttonsContainer(w fyne.Window, boundSource binding.String, appServices *services.AppServices) fyne.CanvasObject {
	scanner := appServices.Scanner
	exporter := appServices.Exporter
	downloader := appServices.Downloader

	exportButton := ExportButton(w, scanner, exporter)
	exportButton.Hide()

	downloadButton := widget.NewButton("Download Progs", func() {
		if err := downloader.Download(); err == nil {
			srcDir, _ := downloader.BoundSourceDir.Get()
			go func() {
				scanner.Scan(srcDir)
			}()
		} else {
			println(err.Error())
		}
	})

	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan, _ := boundSource.Get()
		go func() {
			scanner.Scan(dirToScan)
		}()
	})

	downloader.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := downloader.IsDownloading.Get()
		if isDownloading {
			scanButton.Disable()
			downloadButton.Disable()
		} else {
			scanButton.Enable()
			downloadButton.Enable()
		}
	}))

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		isScanning, _ := scanner.IsScanning.Get()
		if isScanning {
			scanButton.Disable()
			downloadButton.Disable()
		} else {
			downloadButton.Enable()
			stories, _ := scanner.BoundStories.Get()
			if len(stories) == 0 {
				scanButton.Enable()
			} else {
				scanButton.Hide()
				exportButton.Show()
			}
		}
	}))

	return container.NewVBox(
		downloadButton,
		scanButton,
		exportButton,
	)
}
