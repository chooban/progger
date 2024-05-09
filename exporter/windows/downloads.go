package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
)

func NewDownloads(a *app.ProggerApp) fyne.CanvasObject {
	c := container.NewBorder(
		nil,
		downloadButton(a.State),
		nil,
		nil,
	)

	return c
}
func downloadButton(d Dispatcher) *widget.Button {
	downloadButton := widget.NewButton("Download Progs", func() {
		d.Dispatch(app.StartDownloadingMessage{})
	})

	return downloadButton
}
