package exporter

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
)

func MainWindow(a fyne.App, w fyne.Window) fyne.CanvasObject {
	// We'll need a scanner service-like object to perform the operations
	scanner := NewScanner()
	scannerPanel := scannerPanel(a, w, scanner)

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, err := scanner.IsScanning.Get(); err != nil {
			println(err.Error())
		} else {
			if !isScanning && len(scanner.Issues) > 0 {
				scannerPanel.Hide()
			}
		}
	}))

	return container.NewBorder(
		widget.NewLabel("Borag Thungg!"),
		nil,
		nil,
		nil,
	)
}

func scannerPanel(a fyne.App, w fyne.Window, scanner *Scanner) fyne.CanvasObject {
	// We want to be able to react to the source directory changing
	boundSource := BoundSourceDir(a)

	dirButton := widget.NewButton("Choose directory", func() {
		dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
			boundSource.Set(l.Path())
		}, w)
	})

	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan, _ := boundSource.Get()
		scanner.Scan(dirToScan)
	})

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		isScanning, _ := scanner.IsScanning.Get()
		if isScanning {
			dirButton.Disable()
			scanButton.Disable()
		} else {
			dirButton.Enable()
			scanButton.Enable()
		}
	}))

	return container.NewBorder(
		widget.NewLabel("Borag Thungg!"),
		container.NewVBox(
			dirButton,
			scanButton,
		),
		nil,
		nil,
		container.NewCenter(
			widget.NewLabelWithData(boundSource),
		),
	)
}

type Scanner struct {
	IsScanning binding.Bool
	Issues     []api.Issue
}

func (s *Scanner) Scan(path string) {
	s.IsScanning.Set(true)
	ctx := WithLogger()
	issues := scan.Dir(ctx, path, 20)
	s.Issues = issues
	if err := s.IsScanning.Set(false); err != nil {
		println(err.Error())
	}
}

func NewScanner() *Scanner {
	isScanning := binding.NewBool()

	return &Scanner{
		IsScanning: isScanning,
	}
}
