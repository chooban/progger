package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
	"slices"
	"strings"
)

type Dispatcher interface {
	Dispatch(msg interface{})
}

func TabWindow(a *app.ProggerApp) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Stories", theme.DocumentIcon(), MainWindow(a)),
		container.NewTabItemWithIcon("Downloads", theme.DownloadIcon(), NewDownloads(a)),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), NewSettingsCanvas(a)),
		container.NewTabItemWithIcon("About", theme.HomeIcon(), NewAbout(a)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}

func MainWindow(a *app.ProggerApp) fyne.CanvasObject {
	// TODO: This is not usefully bound
	//boundSource := binding.NewString()
	scannerButtonsPanel := buttonsContainer(a)
	displayPanel := displayContainer(a)

	return container.NewBorder(
		nil,
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

func displayContainer(a *app.ProggerApp) fyne.CanvasObject {
	scannerProgressContainer := newScannerContainer()
	downloadProgressContainer := newDownloadProgressContainer()
	sourceDirectoryLabel := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithData(a.AppService.Prefs.BoundSourceDir),
		),
	)

	listContainer := container.NewStack(newStoryListWidget(a.State.Stories))
	listContainer.Hide()

	layout := container.NewStack(
		sourceDirectoryLabel,
		scannerProgressContainer,
		downloadProgressContainer,
		listContainer,
	)

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		if isDownloading, _ := a.State.IsDownloading.Get(); isDownloading == true {
			sourceDirectoryLabel.Hide()
			scannerProgressContainer.Hide()
			downloadProgressContainer.Show()
		} else {
			sourceDirectoryLabel.Show()
			scannerProgressContainer.Show()
			downloadProgressContainer.Hide()
		}
	}))

	a.State.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := a.State.IsScanning.Get(); isScanning == true {
			sourceDirectoryLabel.Hide()
			downloadProgressContainer.Hide()
			scannerProgressContainer.Show()
		} else {
			stories, err := a.State.Stories.Get()

			if err != nil {
				println(err.Error())
				return
			}
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

func buttonsContainer(a *app.ProggerApp) fyne.CanvasObject {
	exportButton := ExportButton(a)
	exportButton.Hide()

	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan := a.AppService.Prefs.SourceDirectory()
		a.State.Dispatch(app.StartScanningMessage{Directory: dirToScan})
	})

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := a.State.IsDownloading.Get()
		if isDownloading {
			scanButton.Disable()
		} else {
			scanButton.Enable()
		}
	}))

	a.State.IsScanning.AddListener(binding.NewDataListener(func() {
		isScanning, _ := a.State.IsScanning.Get()
		if isScanning {
			scanButton.Disable()
		} else {
			stories, _ := a.State.Stories.Get()
			if len(stories) == 0 {
				scanButton.Enable()
			} else {
				scanButton.Hide()
				exportButton.Show()
			}
		}
	}))

	return container.NewVBox(
		scanButton,
		exportButton,
	)
}

func ExportButton(a *app.ProggerApp) *widget.Button {
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
					if err := exporter.Export(toExport, prefsService.SourceDirectory(), prefsService.ExportDirectory(), fname); err != nil {
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
