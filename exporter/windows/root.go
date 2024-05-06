package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
	"github.com/chooban/progger/exporter/prefs"
	"github.com/chooban/progger/exporter/services"
	"slices"
	"strings"
)

func MainWindow(app *app.ProggerApp) fyne.CanvasObject {
	boundSource := prefs.BoundSourceDir(app.FyneApp)

	scannerButtonsPanel := buttonsContainer(app.RootWindow, boundSource, app.AppService)
	displayPanel := displayContainer(boundSource, app.AppService)

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
	downloadProgressContainer := newDownloadProgressContainer()
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
		downloadProgressContainer,
		listContainer,
	)

	downloader.IsDownloading.AddListener(binding.NewDataListener(func() {
		if isDownloading, _ := downloader.IsDownloading.Get(); isDownloading == true {
			sourceDirectoryLabel.Hide()
			scannerProgressContainer.Hide()
			downloadProgressContainer.Show()
		} else {
			sourceDirectoryLabel.Show()
			scannerProgressContainer.Show()
			downloadProgressContainer.Hide()
		}
	}))

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := scanner.IsScanning.Get(); isScanning == true {
			sourceDirectoryLabel.Hide()
			downloadProgressContainer.Hide()
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

func buttonsContainer(w fyne.Window, boundSource binding.String, appServices *services.AppServices) fyne.CanvasObject {
	scanner := appServices.Scanner
	exporter := appServices.Exporter
	downloader := appServices.Downloader

	exportButton := ExportButton(w, scanner, exporter)
	exportButton.Hide()

	downloadButton := DownloadButton(w, appServices.Downloader, appServices.Scanner)
	scanButton := widget.NewButton("Scan Directory", func() {
		println("Scan button clicked")
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

func DownloadButton(w fyne.Window, downloader *services.Downloader, scanner *services.Scanner) *widget.Button {
	downloadButton := widget.NewButton("Download Progs", func() {
		println("Starting downloads")
		if err := downloader.Download(); err == nil {
			println("Finished downloading")
			srcDir, _ := downloader.BoundSourceDir.Get()
			scanner.Scan(srcDir)
		} else {
			println(err.Error())
		}
	})

	return downloadButton
}

func ExportButton(w fyne.Window, scanner *services.Scanner, exporter *services.Exporter) *widget.Button {
	exportButton := widget.NewButton("Export Story", func() {
		stories, err := scanner.BoundStories.Get()
		if err != nil {
			dialog.ShowError(err, w)
		}
		toExport := make([]*api.Story, 0)
		for _, v := range stories {
			story := v.(*api.Story)
			if story.ToExport {
				toExport = append(toExport, story)
			}
		}
		if len(toExport) == 0 {
			dialog.ShowInformation("Export", "No stories selected", w)
		} else {
			filename := binding.NewString()
			filename.Set(toExport[0].Display() + ".pdf")
			fnameEntry := widget.NewEntryWithData(filename)

			onClose := func(b bool) {
				if b {
					fname, _ := filename.Get()
					if err := exporter.Export(toExport, fname); err != nil {
						dialog.ShowError(err, w)
					} else {
						dialog.ShowInformation("Export", "File successfully exported", w)
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
				w,
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
