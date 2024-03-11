package download

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
	"path/filepath"
)

func browser(ctx context.Context) (playwright.BrowserContext, error) {
	logger := logr.FromContextOrDiscard(ctx)
	pw, err := playwright.Run()
	if err != nil {
		logger.Error(err, "failed to open browser")
		return nil, err
	}
	var configDir string
	if configDir, err = BrowserContextDir(ctx); err != nil || configDir == "" {
		logger.Error(err, "failed to get context dir for browser")
		return nil, err
	}

	contextDir := filepath.Join(configDir, "browser")
	bContext, err := pw.Firefox.LaunchPersistentContext(
		contextDir,
		playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless:          boolPointer(true),
			JavaScriptEnabled: boolPointer(false),
		},
	)
	if err != nil {
		return nil, err
	}

	return bContext, nil
}

func boolPointer(b bool) *bool {
	return &b
}
