package windows

import (
	"bytes"
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/app"
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

func newAboutCanvas(app *app.ProggerApp) fyne.CanvasObject {
	r := bytes.NewReader(icon)
	img, err := png.Decode(r)
	if err != nil {
		println("Cannot load icon for about window")
		println(err.Error())
		panic("Cannot load icon for about window")
	}
	m := app.FyneApp.Metadata()
	d := newAbout(app.FyneApp, aboutProps{
		m.Version, m.Build, img,
	})
	return d
}

type aboutProps struct {
	Version string
	Build   int
	Icon    image.Image
}

func newAbout(app fyne.App, p aboutProps) fyne.CanvasObject {
	iconImage := canvas.NewImageFromImage(p.Icon)
	iconImage.FillMode = canvas.ImageFillOriginal

	heading := widget.NewLabelWithStyle("Progger", fyne.TextAlignCenter, fyne.TextStyle{
		Bold:      true,
		Italic:    false,
		Monospace: false,
		Symbol:    false,
		TabWidth:  0,
	})

	description := widget.NewRichTextFromMarkdown(fmt.Sprintf("Version %s (%d) ([MIT License](https://github.com/chooban/progger/blob/main/LICENSE))", p.Version, p.Build))

	licenses := widget.NewRichTextFromMarkdown(licencesText + extraLicensesText)
	licenses.Wrapping = fyne.TextWrapWord

	licensesContainer := container.NewVScroll(licenses)

	l := container.NewBorder(
		container.NewVBox(iconImage, heading, description),
		nil,
		nil,
		nil,
		licensesContainer,
	)

	return l
}
