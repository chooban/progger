package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
)

type Dispatcher interface {
	Dispatch(msg interface{})
}

func newDownloadsCanvas(a *app.ProggerApp) fyne.CanvasObject {
	progress := downloadProgress()
	dButton := downloadButton(a.State)

	mainPanel := container.NewStack(
		progress,
		dButton,
	)

	centeredPanel := container.NewCenter(mainPanel)

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := a.State.IsDownloading.Get()

		if isDownloading {
			showHide(mainPanel, progress)
		} else {
			showHide(mainPanel, dButton)
		}
	}))

	c := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		centeredPanel,
	)

	return c
}

func downloadProgress() *fyne.Container {
	mainPanel := container.New(
		layout.NewVBoxLayout(),
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)

	return mainPanel
}

func downloadButton(d Dispatcher) *widget.Button {
	downloadButton := widget.NewButton("Download Prog List", func() {
		d.Dispatch(app.StartDownloadingProgListMessage{})
	})

	return downloadButton
}
