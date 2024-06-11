package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/chooban/progger/exporter/app"
)

func TabWindow(a *app.ProggerApp) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Stories", theme.DocumentIcon(), newStoriesCanvas(a)),
		container.NewTabItemWithIcon("Downloads", theme.DownloadIcon(), newDownloadsCanvas(a)),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), newSettingsCanvas(a)),
		container.NewTabItemWithIcon("About", theme.HomeIcon(), newAboutCanvas(a)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
