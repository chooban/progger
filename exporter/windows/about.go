package windows

import (
	"bytes"
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"image"
	"image/png"
)

//go:embed Icon.png
var icon []byte

//go:embed extra-licenses.txt
var extraLicensesText string

var licencesText = `# Open Source Licenses

Icon provided by [Atlas Icons](https://atlasicons.vectopus.com/) ([MIT License](https://github.com/Vectopus/Atlas-icons-font/blob/main/LICENSE))

This software uses a variety of Open Source Software to run.

* [PDFium](https://github.com/chromium/pdfium) ([Apache License 2.0](https://github.com/chromium/pdfium/blob/main/LICENSE))
`

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
	w.Resize(fyne.Size{Width: 400, Height: 400})

	iconImage := canvas.NewImageFromImage(p.Icon)
	iconImage.FillMode = canvas.ImageFillOriginal

	heading := widget.NewLabelWithStyle("Progger", fyne.TextAlignCenter, fyne.TextStyle{
		Bold:      true,
		Italic:    false,
		Monospace: false,
		Symbol:    false,
		TabWidth:  0,
	})

	description := widget.NewRichTextFromMarkdown(fmt.Sprintf("Version %s ([MIT License](https://github.com/chooban/progger/blob/main/LICENSE))", p.Version))

	licenses := widget.NewRichTextFromMarkdown(licencesText + extraLicensesText)
	licenses.Wrapping = fyne.TextWrapWord

	licensesContainer := container.NewVScroll(licenses)

	l := container.NewBorder(
		container.NewVBox(iconImage, heading, description),
		widget.NewButton("Close", func() {
			w.Close()
		}),
		nil,
		nil,
		licensesContainer,
	)
	//c := container.New(l)
	w.SetContent(l)

	return w
}
