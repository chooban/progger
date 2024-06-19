package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	downloadApi "github.com/chooban/progger/download/api"
	"github.com/chooban/progger/exporter/app"
	"time"
)

type Dispatcher interface {
	Dispatch(msg interface{})
}

func newDownloadsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	progress := downloadProgress()
	dButton := downloadButton(a.State)
	progListWidget := progList(a.State.AvailableProgs, a.State)

	mainPanel := container.New(
		layout.NewStackLayout(),
		progress,
		dButton,
		progListWidget,
	)

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := a.State.IsDownloading.Get()
		if isDownloading {
			showHide(mainPanel, progress)
		} else {
			availableProgs, _ := a.State.AvailableProgs.Get()
			if len(availableProgs) > 0 {
				showHide(mainPanel, progListWidget)
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

func progList(progs binding.UntypedList, d Dispatcher) fyne.CanvasObject {
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
			// ideally we should check `ok` for each one of those casting
			// but we know that they are those types for sure
			label := ctr.Objects[0].(*widget.Label)
			//check := ctr.Objects[1].(*widget.Check)
			diu, _ := di.(binding.Untyped).Get()
			prog := diu.(downloadApi.DigitalComic)

			//b := binding.BindBool(&story.ToExport)
			//check.Bind(b)
			progDate, _ := time.Parse("2006-01-02", prog.IssueDate)
			formattedDate := formatDateWithOrdinal(progDate)
			label.SetText(fmt.Sprintf("Prog %d (%s)", prog.IssueNumber, formattedDate))
		},
	)

	c := container.NewBorder(
		nil, nil, nil, nil, listOfProgs,
	)

	return c
}

func downloadProgress() *fyne.Container {
	mainPanel := container.New(
		layout.NewVBoxLayout(),
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)

	return container.NewCenter(mainPanel)
}

func downloadButton(d Dispatcher) fyne.CanvasObject {
	downloadButton := widget.NewButton("Download Prog List", func() {
		d.Dispatch(app.StartDownloadingProgListMessage{})
	})

	return container.NewCenter(downloadButton)
}

// formatDateWithOrdinal prints a given time in the format 1st January 2000.
func formatDateWithOrdinal(t time.Time) string {
	return fmt.Sprintf("%s %s %d", addOrdinal(t.Day()), t.Month(), t.Year())
}

// addOrdinal takes a number and adds its ordinal (like st or th) to the end.
func addOrdinal(n int) string {
	switch n {
	case 1, 21, 31:
		return fmt.Sprintf("%dst", n)
	case 2, 22:
		return fmt.Sprintf("%dnd", n)
	case 3, 23:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
