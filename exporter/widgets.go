package exporter

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/services"
)

func DownloadButton(w fyne.Window, downloader *services.Downloader, scanner *services.Scanner) *widget.Button {
	downloadButton := widget.NewButton("Download Progs", func() {
		if err := downloader.Download(); err == nil {
			srcDir, _ := downloader.BoundSourceDir.Get()
			scanner.Scan(srcDir)
		} else {
			println(err.Error())
		}
	})

	return downloadButton
}

func ExportButton(w fyne.Window, scanner *services.Scanner, exporter *services.Exporter) *widget.Button {
	exportButton := widget.NewButton("Export Story", func() {
		stories, err := scanner.BoundStories.Get()
		if err != nil {
			dialog.ShowError(err, w)
		}
		toExport := make([]*api.Story, 0)
		for _, v := range stories {
			story := v.(*api.Story)
			if story.ToExport {
				toExport = append(toExport, story)
			}
		}
		if len(toExport) == 0 {
			dialog.ShowInformation("Export", "No stories selected", w)
		} else {
			filename := binding.NewString()
			filename.Set(toExport[0].Display() + ".pdf")
			fnameEntry := widget.NewEntryWithData(filename)

			onClose := func(b bool) {
				if b {
					fname, _ := filename.Get()
					if err := exporter.Export(toExport, fname); err != nil {
						dialog.ShowError(err, w)
					} else {
						dialog.ShowInformation("Export", "File successfully exported", w)
					}
				}
			}

			formDialog := dialog.NewForm(
				"Export",
				"Export",
				"Cancel",
				[]*widget.FormItem{
					{Text: "Filename", Widget: fnameEntry},
				},
				onClose,
				w,
			)
			formDialog.Show()
			formDialog.Resize(fyne.NewSize(500, 100))
		}
	})

	return exportButton
}
