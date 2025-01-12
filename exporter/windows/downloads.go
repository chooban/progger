package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
	"reflect"
)

type Dispatcher interface {
	Dispatch(msg interface{})
}

func newDownloadsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	//downloadableCandidates := binding.NewUntypedList()

	//downloadableCandidates := a.State.AvailableProgs

	progress := newDownloadProgress()
	dButton := downloadButton(a.State, "Download Prog List")
	progListContainer := newProgListContainer(a.State.AvailableProgs, a.State)

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

func newProgListContainer(progs binding.UntypedList, d Dispatcher) *fyne.Container {
	progListWidget := newProgList(progs)
	refreshDownloadListButton := widget.NewButton("Refresh Prog List", func() {
		d.Dispatch(app.StartDownloadingProgListMessage{})
	})
	downloadAllButton := widget.NewButton("Download Latest Prog", func() {
		d.Dispatch(app.DownloadLatestMessage{})
	})

	nothingToDownload := widget.NewLabel("No new progs to download")
	mainDisplay := container.NewStack(progListWidget, nothingToDownload)
	progListContainer := container.NewBorder(nil, container.NewCenter(
		container.NewHBox(refreshDownloadListButton, downloadAllButton)), nil, nil, mainDisplay)

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

			labelText := prog.Comic.String()
			check := ctr.Objects[1].(*widget.Check)

			if prog.Downloaded {
				check.SetChecked(true)
				check.Disable()
			} else {
				check.SetChecked(false)
				check.Enable()
			}

			// TODO: use check.onChanged to toggle whether or not we should download this prog when requested.
			// Since this information is state and should persist between window changes, we'll need to maintain
			// the list outside of this component

			if reflect.TypeOf(ctr.Objects[1]).String() == "*widget.Label" {
				label := ctr.Objects[0].(*widget.Label)
				label.SetText(prog.Comic.String())
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
