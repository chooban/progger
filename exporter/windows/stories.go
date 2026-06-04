package windows

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/chooban/progger/database"
	"github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/exporter/app"
)

func showHide(container *fyne.Container, toShow fyne.CanvasObject) {
	for i := 0; i < len(container.Objects); i++ {
		container.Objects[i].Hide()
	}
	toShow.Show()
}

func newStoriesCanvas(a *app.ProggerApp) fyne.CanvasObject {
	storiesPanel := storiesContainer(a)
	scannerProgress := newScannerProgressContainer()
	downloadProgress := newDownloadProgressContainer()

	centralLayout := container.New(
		layout.NewStackLayout(),
		storiesPanel,
		scannerProgress,
		downloadProgress,
	)

	a.State.IsScanning.AddListener(binding.NewDataListener(func() {
		if isScanning, _ := a.State.IsScanning.Get(); isScanning == true {
			showHide(centralLayout, scannerProgress)
		} else {
			showHide(centralLayout, storiesPanel)
		}
	}))

	a.State.IsDownloading.AddListener(binding.NewDataListener(func() {
		if isDownloading, _ := a.State.IsDownloading.Get(); isDownloading == true {
			println("Showing the download progress")
			showHide(centralLayout, downloadProgress)
		} else {
			showHide(centralLayout, storiesPanel)
		}
	}))

	storiesLayout := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		centralLayout,
	)

	if isDownloading, _ := a.State.IsDownloading.Get(); isDownloading == true {
		showHide(centralLayout, downloadProgress)
	} else {
		showHide(centralLayout, storiesPanel)
	}

	return storiesLayout
}

func newScannerProgressContainer() *fyne.Container {
	barContainer := container.NewVBox(
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Scanning..."),
	)
	centeredBar := container.NewCenter(
		barContainer,
	)

	return centeredBar
}

func newStoryListWidget(boundStories binding.UntypedList) *fyne.Container {
	filterValue := binding.NewString()
	filteredList := binding.NewUntypedList()

	storyList := widget.NewListWithData(
		filteredList,
		// Component structure of the row
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewCheck("", func(b bool) {}),
				// takes the rest of the space
				widget.NewLabel(""),
			)
		},
		func(di binding.DataItem, o fyne.CanvasObject) {
			ctr, _ := o.(*fyne.Container)
			// ideally we should check `ok` for each one of those casting
			// but we know that they are those types for sure
			label := ctr.Objects[0].(*widget.Label)
			check := ctr.Objects[1].(*widget.Check)
			diu, _ := di.(binding.Untyped).Get()
			story := diu.(*api.Story)

			b := binding.BindBool(&story.ToExport)
			label.SetText(fmt.Sprintf("%s (%s)", story.Display(), story.IssueSummary()))
			check.Bind(b)
		},
	)

	listRefresh := func() {
		stories, _ := boundStories.Get()
		toDisplay := make([]interface{}, 0, len(stories))

		f, _ := filterValue.Get()
		if strings.TrimSpace(f) == "" {
			toDisplay = slices.Clone(stories)
		} else {
			_f := strings.Split(strings.ToLower(f), " ")
			for _, v := range stories {
				if ContainsAll(strings.ToLower(v.(*api.Story).Display()), _f) {
					toDisplay = append(toDisplay, v)
				}
			}
		}

		if err := filteredList.Set(toDisplay); err != nil {
			println(err.Error())
		}
		storyList.Refresh()
	}

	boundStories.AddListener(binding.NewDataListener(listRefresh))
	filterValue.AddListener(binding.NewDataListener(listRefresh))

	filter := widget.NewEntryWithData(filterValue)
	filter.SetPlaceHolder("Filter list")

	c := container.NewBorder(
		nil,
		filter,
		nil,
		nil,
		storyList,
	)

	return c
}

func noStoriesContainer(a *app.ProggerApp) fyne.CanvasObject {
	scanButton := widget.NewButton("Scan Directory", func() {
		startScan(a)
	})
	content := container.NewVBox(widget.NewLabelWithData(a.Services.Prefs.ProgSourceDir), scanButton)

	contentWrapper := container.NewCenter(
		content,
	)

	return contentWrapper
}

func startScan(a *app.ProggerApp) {
	dirsToScan := []string{a.Services.Prefs.ProgSourceDirectory(), a.Services.Prefs.MegSourceDirectory()}
	knownTitles := readKnownTitles(a)
	skipTitles := readSkipTitles(a)

	// Create the operation
	op := app.NewScanOperation()

	ctx, cancel := context.WithCancel(context.Background())
	op.SetCancel(cancel)

	go func() {
		_ = op.IsRunning.Set(true)
		defer func() {
			_ = op.IsRunning.Set(false)
		}()

		foundStories, err := a.Services.Scanner.Scan(ctx, dirsToScan, knownTitles, skipTitles)
		if err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Convert to untyped for binding
		untypedStories := make([]interface{}, len(foundStories))
		storiesToStore := make([]api.Story, len(foundStories))
		for i, v := range foundStories {
			untypedStories[i] = v
			storiesToStore[i] = *v
		}

		if err := op.Stories.Set(untypedStories); err != nil {
			_ = op.Error.Set(err.Error())
			return
		}

		// Persist stories to database
		persistStories(a, storiesToStore, "2000 AD")
	}()

	// Bind the operation state to the app state
	op.IsRunning.AddListener(binding.NewDataListener(func() {
		isRunning, _ := op.IsRunning.Get()
		a.State.IsScanning.Set(isRunning)
	}))

	op.Stories.AddListener(binding.NewDataListener(func() {
		stories, _ := op.Stories.Get()
		a.State.Stories.Set(stories)
	}))
}

func storiesContainer(a *app.ProggerApp) fyne.CanvasObject {
	listContainer := container.NewBorder(
		nil, storiesButtonsContainer(a), nil, nil,
		newStoryListWidget(a.State.Stories),
	)
	noListContainer := noStoriesContainer(a)

	c := container.New(layout.NewStackLayout(), listContainer, noListContainer)

	boundStories := a.State.Stories
	boundStories.AddListener(binding.NewDataListener(func() {
		stories, _ := boundStories.Get()
		if len(stories) == 0 {
			showHide(c, noListContainer)
		} else {
			showHide(c, listContainer)
		}
	}))

	return c
}

func newDownloadProgressContainer() *fyne.Container {
	barContainer := container.NewVBox(
		widget.NewProgressBarInfinite(),
		widget.NewLabel("Downloading..."),
	)
	centeredBar := container.NewCenter(
		barContainer,
	)

	return centeredBar
}

func storiesButtonsContainer(a *app.ProggerApp) fyne.CanvasObject {
	exportButton := exportButton(a)

	scanButton := widget.NewButton("Force Rescan", func() {
		startScan(a)
	})

	return container.NewVBox(
		exportButton,
		scanButton,
	)
}

func exportButton(a *app.ProggerApp) *widget.Button {
	exporter := a.Services.Exporter
	prefsService := a.Services.Prefs

	exportButton := widget.NewButton("Export Story", func() {
		stories, err := a.State.Stories.Get()
		if err != nil {
			dialog.ShowError(err, a.RootWindow)
		}
		toExport := make([]*api.Story, 0)
		for _, v := range stories {
			story := v.(*api.Story)
			if story.ToExport {
				toExport = append(toExport, story)
			}
		}
		if len(toExport) == 0 {
			dialog.ShowInformation("Export", "No stories selected", a.RootWindow)
		} else {
			filename := binding.NewString()
			filename.Set(toExport[0].Display() + ".pdf")
			fnameEntry := widget.NewEntryWithData(filename)

			artistBool := binding.NewBool()
			artistCheckbox := widget.NewCheckWithData("", artistBool)

			artistCheckbox.OnChanged = func(v bool) {
				_f, _ := filename.Get()
				if v {
					_f = strings.TrimSuffix(_f, ".pdf") + " - Artists Edition.pdf"
				} else {
					_f = strings.TrimSuffix(_f, " - Artists Edition.pdf") + ".pdf"
				}
				filename.Set(_f)
				artistBool.Set(v)
			}

			onClose := func(b bool) {
				if b {
					fname, _ := filename.Get()
					exportArtistEd, _ := artistBool.Get()

					ctx, _, _ := app.WithLogger()
					if err := exporter.Export(ctx, toExport, exportArtistEd, prefsService.ExportDirectory(), fname); err != nil {
						dialog.ShowError(err, a.RootWindow)
					} else {
						dialog.ShowInformation("Export", "File successfully exported", a.RootWindow)
					}
				}
			}

			formDialog := dialog.NewForm(
				"Export",
				"Export",
				"Cancel",
				[]*widget.FormItem{
					{Text: "Filename", Widget: fnameEntry},
					{Text: "Artists Edition", Widget: artistCheckbox},
				},
				onClose,
				a.RootWindow,
			)
			formDialog.Show()
			formDialog.Resize(fyne.NewSize(500, 100))
		}
	})

	return exportButton
}

func ContainsAll(s string, t []string) bool {
	if len(t) == 0 {
		return true
	}
	for _, v := range t {
		if !strings.Contains(s, v) {
			return false
		}
	}
	return true
}

func readKnownTitles(a *app.ProggerApp) []string {
	if a.Services.DB != nil {
		repo := database.NewKnownTitlesRepo(a.Services.DB)
		titles, err := repo.List(context.Background())
		if err == nil {
			return titles
		}
	}
	return nil
}

func readSkipTitles(a *app.ProggerApp) []string {
	if a.Services.DB != nil {
		repo := database.NewSkipTitlesRepo(a.Services.DB)
		titles, err := repo.List(context.Background())
		if err == nil {
			return titles
		}
	}
	return nil
}

func persistStories(a *app.ProggerApp, stories []api.Story, publication string) {
	if a.Services.DB == nil {
		return
	}

	ctx := context.Background()
	db := a.Services.DB

	libraryRepo := database.NewLibraryRepo(db)
	seriesRepo := database.NewSeriesRepo(db)
	bookRepo := database.NewBookRepo(db)

	libName := "2000 AD"
	lib, err := libraryRepo.EnsureLibrary(ctx, libName)
	if err != nil {
		println("failed to ensure library:", err.Error())
		return
	}

	for _, story := range stories {
		seriesID := database.HashEntityID("series", story.Series)
		series := &database.Series{
			ID:        seriesID,
			LibraryID: lib.ID,
			Name:      story.Series,
		}
		if err := seriesRepo.Upsert(ctx, series); err != nil {
			println("failed to upsert series:", err.Error())
			continue
		}

		bookID := database.HashEntityID("book", story.Series, story.Title)
		book := &database.Book{
			ID:          bookID,
			SeriesID:    seriesID,
			Name:        story.Title,
			Publication: publication,
			Status:      "READY",
		}
		if err := bookRepo.Upsert(ctx, book); err != nil {
			println("failed to upsert book:", err.Error())
			continue
		}

		var episodes []*database.Episode
		for _, ep := range story.Episodes {
			episode := &database.Episode{
				ID:          database.HashEntityID("episode", story.Series, story.Title, fmt.Sprintf("%d-%d", ep.IssueNumber, ep.Episode.Part)),
				BookID:      bookID,
				Filename:    ep.Filename,
				IssueNumber: ep.IssueNumber,
				Title:       ep.Episode.Title,
				Part:        ep.Episode.Part,
				PageFrom:    ep.Episode.FirstPage,
				PageTo:      ep.Episode.LastPage,
			}
			episodes = append(episodes, episode)
		}
		if err := bookRepo.UpsertEpisodes(ctx, episodes); err != nil {
			println("failed to upsert episodes:", err.Error())
		}
	}

	runExportPostScanSQL(ctx, db)
}

func runExportPostScanSQL(ctx context.Context, db *database.DB) {
	db.ExecContext(ctx, `
with book_counts as (
	select series_id, count(*) as c
	from books b
	group by series_id
	)
update series set book_count = ( select c from book_counts where book_counts.series_id = series.id )
`)

	db.ExecContext(ctx, `
with page_counts as (
	select book_id, sum((e.page_to - e.page_from) + 1) as c
	from episodes e
	group by 1
)
update books set page_count = (
	select c from page_counts where page_counts.book_id = books.id
)
`)

	db.ExecContext(ctx, `
with first_issues as (
select distinct
	books.id,
	first_value(episodes.issue_number) OVER (partition by books.id ORDER BY issue_number) as issue_number
from books
join episodes on (episodes.book_id = books.id)
),
last_issues as (
	select distinct
		books.id,
		first_value(episodes.issue_number) OVER (partition by books.id ORDER BY issue_number DESC) as issue_number
	from books
	join episodes on (episodes.book_id = books.id)
)
update books
set first_issue = (select issue_number from first_issues where id = books.id), last_issue = (select issue_number from last_issues where id = books.id)
`)

	db.ExecContext(ctx, `
with book_order as (
	select 
		books.id, 
		row_number() OVER (PARTITION BY series_id ORDER BY first_issue) as row_number
	from books 
	order by id ASC
)
update books set number = ( select row_number from book_order where id = books.id limit 1 )
`)
}
