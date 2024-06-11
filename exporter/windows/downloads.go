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
	mainPanel := container.New(
		layout.NewVBoxLayout(),
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)
	mainPanel.Hide()
	centeredPanel := container.NewCenter(mainPanel)

	dButton := downloadButton(a.State)

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		isDownloading, _ := a.State.IsDownloading.Get()

		if isDownloading {
			mainPanel.Show()
			dButton.Disable()
		} else {
			mainPanel.Hide()
			dButton.Enable()
		}
	}))

	c := container.NewBorder(
		nil,
		dButton,
		nil,
		nil,
		centeredPanel,
	)

	return c
}
func downloadButton(d Dispatcher) *widget.Button {
	downloadButton := widget.NewButton("Download Progs", func() {
		d.Dispatch(app.StartDownloadingMessage{})
	})

	return downloadButton
}
