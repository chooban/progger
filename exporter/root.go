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
	// We want to be able to react to the source directory changing
	boundSource := BoundSourceDir(a)

	// We'll need a scanner service-like object to perform the operations
	scanner := NewScanner()

	scannerButtonsPanel := buttonsContainer(w, boundSource, scanner)
	displayPanel := displayContainer(w, boundSource, scanner)

	return container.NewBorder(
		widget.NewLabel("Borag Thungg!"),
		scannerButtonsPanel,
		nil,
		nil,
		displayPanel,
	)
}

func displayContainer(w fyne.Window, boundSource binding.String, scanner *Scanner) fyne.CanvasObject {
	barContainer := container.NewVBox(
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Scanning..."),
	)

	label := widget.NewLabelWithData(boundSource)

	centeredLabel := container.NewCenter(label)
	centeredBar := container.NewCenter(
		barContainer,
	)

	layout := container.NewStack(
		centeredLabel,
		centeredBar,
	)
	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := scanner.IsScanning.Get(); isScanning == true {
			centeredLabel.Hide()
			centeredBar.Show()
		} else {
			centeredLabel.Show()
			centeredBar.Hide()
		}
	}))

	return layout
}

func buttonsContainer(w fyne.Window, boundSource binding.String, scanner *Scanner) fyne.CanvasObject {

	dirButton := widget.NewButton("Choose directory", func() {
		dialog.ShowFolderOpen(func(l fyne.ListableURI, err error) {
			boundSource.Set(l.Path())
		}, w)
	})

	scanButton := widget.NewButton("Scan Directory", func() {
		dirToScan, _ := boundSource.Get()
		scanner.Scan(dirToScan)
	})

	exportButton := widget.NewButton("Export Story", func() {})
	exportButton.Hide()
	exportButton.Disable()

	scanner.IsScanning.AddListener(binding.NewDataListener(func() {
		isScanning, _ := scanner.IsScanning.Get()
		if isScanning {
			dirButton.Disable()
			scanButton.Disable()
		} else {
			if len(scanner.Issues) == 0 {
				dirButton.Enable()
				scanButton.Enable()
			} else {
				dirButton.Hide()
				scanButton.Hide()
				exportButton.Show()
			}
		}
	}))

	return container.NewVBox(
		dirButton,
		scanButton,
		exportButton,
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
