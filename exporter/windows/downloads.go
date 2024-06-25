package windows

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	downloadApi "github.com/chooban/progger/download/api"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

type Dispatcher interface {
	Dispatch(msg interface{})
}

func newDownloadsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	downloadableCandidates := binding.NewUntypedList()

	progress := newDownloadProgress()
	dButton := downloadButton(a.State, "Download Prog List")
	progListContainer := newProgListContainer(downloadableCandidates, a.State)

	mainPanel := container.New(
		layout.NewStackLayout(),
		progress,
		dButton,
		progListContainer,
	)

	refreshDownloadableProgs := func() {
		downloads, _ := a.State.AvailableProgs.Get()
		if len(downloads) > 0 {
			candidates := make([]interface{}, 0, len(downloads))
			for _, v := range downloads {
				p := v.(api.Downloadable)
				maybePath := filepath.Join(a.AppService.Prefs.SourceDirectory(), p.Comic.Filename(downloadApi.Pdf))
				_, err := os.Stat(maybePath)
				if errors.Is(err, os.ErrNotExist) {
					candidates = append(candidates, p)
				}
			}
			println(fmt.Sprintf("Setting %d download candidates", len(candidates)))
			downloadableCandidates.Set(candidates)
		}
	}

	a.State.AvailableProgs.AddListener(binding.NewDataListener(refreshDownloadableProgs))
	a.AppService.Prefs.BoundSourceDir.AddListener(binding.NewDataListener(refreshDownloadableProgs))

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

func newProgListContainer(progs binding.UntypedList, d Dispatcher) *fyne.Container {
	progListWidget := newProgList(progs)
	downloadButton := widget.NewButton("Re-download Prog List", func() {
		d.Dispatch(app.StartDownloadingProgListMessage{})
	})

	nothingToDownload := widget.NewLabel("No new progs to download")
	mainDisplay := container.NewStack(progListWidget, nothingToDownload)

	progListContainer := container.NewBorder(nil, container.NewCenter(downloadButton), nil, nil, mainDisplay)

	progs.AddListener(binding.NewDataListener(func() {
		downloadCandidates, _ := progs.Get()
		if len(downloadCandidates) > 0 {
			showHide(mainDisplay, progListWidget)
		} else {
			showHide(mainDisplay, nothingToDownload)
		}
	}))

	return progListContainer
}

func newProgList(progs binding.UntypedList) fyne.CanvasObject {
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
			prog := diu.(api.Downloadable)

			progDate, _ := time.Parse("2006-01-02", prog.Comic.IssueDate)
			formattedDate := formatDateWithOrdinal(progDate)

			labelText := fmt.Sprintf("Prog %d (%s)", prog.Comic.IssueNumber, formattedDate)
			check := ctr.Objects[1].(*widget.Check)
			check.SetChecked(false)
			check.Enable()

			if reflect.TypeOf(ctr.Objects[1]).String() == "*widget.Label" {
				label := ctr.Objects[0].(*widget.Label)
				label.SetText(fmt.Sprintf("Prog %d (%s)", prog.Comic.IssueNumber, formattedDate))
			} else {
				label := widget.NewLabel(labelText)
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
	mainPanel := container.New(
		layout.NewVBoxLayout(),
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)

	return container.NewCenter(mainPanel)
}

func downloadButton(d Dispatcher, label string) fyne.CanvasObject {
	downloadButton := widget.NewButton(label, func() {
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
