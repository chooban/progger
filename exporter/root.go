package exporter

import (
	"cmp"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"path/filepath"
	"slices"
)

func MainWindow(a fyne.App, w fyne.Window) fyne.CanvasObject {
	// We want to be able to react to the source directory changing
	boundSource := BoundSourceDir(a)

	// We'll need a scanner service-like object to perform the operations
	scanner := NewScanner()

	scannerButtonsPanel := buttonsContainer(w, boundSource, scanner)
	displayPanel := displayContainer(w, boundSource, scanner)

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

func newStoryListWidget(boundStories binding.UntypedList) *widget.List {
	storyList := widget.NewListWithData(
		boundStories,
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
			story := diu.(*Story)

			b := binding.BindBool(&story.ToExport)
			l.SetText(fmt.Sprintf("%s - %s (%s)", story.Series, story.Title, story.IssueSummary()))
			c.Bind(b)
		},
	)

	return storyList
}

func displayContainer(w fyne.Window, boundSource binding.String, scanner *Scanner) fyne.CanvasObject {

	scannerProgressContainer := newScannerContainer()
	sourceDirectoryLabel := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithData(boundSource),
			widget.NewButton("Choose Directory", func() {
				dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
					boundSource.Set(l.Path())
				}, w)
			}),
		),
	)

	listContainer := container.NewStack(newStoryListWidget(scanner.BoundStories))
	listContainer.Hide()

	layout := container.NewStack(
		sourceDirectoryLabel,
		scannerProgressContainer,
		listContainer,
	)

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

func buttonsContainer(w fyne.Window, boundSource binding.String, scanner *Scanner) fyne.CanvasObject {
	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan, _ := boundSource.Get()
		go func() {
			scanner.Scan(dirToScan)
		}()
	})

	exportButton := widget.NewButton("Export Story", func() {
		stories, err := scanner.BoundStories.Get()
		sourceDir, _ := boundSource.Get()
		if err != nil {
			println(err.Error())
		}
		toExport := make([]api.ExportPage, 0)
		for _, v := range stories {
			story := v.(*Story)
			if story.ToExport {
				for _, e := range story.Episodes {
					toExport = append(toExport, api.ExportPage{
						Filename:    filepath.Join(sourceDir, e.Filename),
						PageFrom:    e.FirstPage,
						PageTo:      e.LastPage,
						IssueNumber: e.IssueNumber,
						Title:       fmt.Sprintf("%s - Part %d", e.Title, e.Part),
					})
				}
			}
		}
		if len(toExport) == 0 {
			println("Nothing to export")
			return
		}
		// Sort by issue number. We sometimes have issues being wrongly grouped, but surely we never want anything
		// other than issue order?
		slices.SortFunc(toExport, func(i, j api.ExportPage) int {
			return cmp.Compare(i.IssueNumber, j.IssueNumber)
		})

		// Do the export
		err = scan.Build(WithLogger(), toExport, "export.pdf")
		if err != nil {
			println(err.Error())
		}
	})
	exportButton.Hide()

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		isScanning, _ := scanner.IsScanning.Get()
		if isScanning {
			scanButton.Disable()
		} else {
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
		scanButton,
		exportButton,
	)
}
