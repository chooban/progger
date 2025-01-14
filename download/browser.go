package download

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
	"os"
	"path/filepath"
	"strconv"
)

func browser(ctx context.Context) (playwright.BrowserContext, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info("Starting to create browser")
	err := playwright.Install()
	if err != nil {
		return nil, err
	}
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
	headless, err := strconv.ParseBool(getEnv("DEBUG", "false"))
	if err != nil {
		headless = false
	}
	timeout := float64(10 * 1000)
	bContext, err := pw.Chromium.LaunchPersistentContext(
		contextDir,
		playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless:          boolPointer(!headless),
			JavaScriptEnabled: boolPointer(false),
			Timeout:           &timeout,
		},
	)
	if err != nil {
		return nil, err
	}

	logger.V(1).Info("Returning browser context")
	return bContext, nil
}

func boolPointer(b bool) *bool {
	return &b
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
