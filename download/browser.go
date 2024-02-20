package download

import (
	"github.com/playwright-community/playwright-go"
	"os"
	"path/filepath"
)

func browser() (playwright.BrowserContext, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	configDir, _ := os.UserConfigDir()
	contextDir := filepath.Join(configDir, "progger", "download", "browser")
	bContext, err := pw.Firefox.LaunchPersistentContext(
		contextDir,
		playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless:          boolPointer(true),
			JavaScriptEnabled: boolPointer(false),
		},
	)
	//bContext, err := pw.Chromium.LaunchPersistentContext(
	//	contextDir,
	//	playwright.BrowserTypeLaunchPersistentContextOptions{Headless: &headless},
	//)
	if err != nil {
		return nil, err
	}

	return bContext, nil
}

func boolPointer(b bool) *bool {
	return &b
}
