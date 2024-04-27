package windows

import (
	"bytes"
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"image"
	"image/png"
)

//go:embed Icon.png
var icon []byte

func NewAbout(app *api.ProggerApp) fyne.Window {
	r := bytes.NewReader(icon)
	img, err := png.Decode(r)
	if err != nil {
		println("Cannot load icon for about window")
		println(err.Error())
		panic("Cannot load icon for about window")
	}

	d := newAbout(app.FyneApp, aboutProps{
		app.FyneApp.Metadata().Version, img,
	})
	return d
}

type aboutProps struct {
	Version string
	Icon    image.Image
}

func newAbout(app fyne.App, p aboutProps) fyne.Window {
	w := app.NewWindow("About")
	w.SetTitle("About")

	iconImage := canvas.NewImageFromImage(p.Icon)
	iconImage.FillMode = canvas.ImageFillOriginal

	heading := widget.NewLabelWithStyle("Progger", fyne.TextAlignCenter, fyne.TextStyle{
		Bold:      true,
		Italic:    false,
		Monospace: false,
		Symbol:    false,
		TabWidth:  0,
	})

	description := widget.NewRichTextFromMarkdown(fmt.Sprintf("Version %s", p.Version))

	c := container.New(
		layout.NewVBoxLayout(),
		iconImage,
		heading,
		description,
	)
	w.SetContent(c)

	return w
}
