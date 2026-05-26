package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabasePath    string
	ScanDirectories []string
	LibraryName     string
	Host            string
	Username        string
	Password        string
	LogLevel        string
	ScanOnStartup   bool
}

var KnownSeries = []string{
	"Anderson, Psi-Division",
	"Chimpsky's Law",
	"Counterfeit Girl",
	"Feral & Foe",
	"Lowborn High",
	"Scarlet Traces",
	"Strontium Dog",
	"Strontium Dug",
	"The Fall of Deadworld",
}

var SkipTitles = []string{
	"Interrogation",
	"New Books",
	"Obituary",
	"Tribute",
	"Untitled",
	"Encyclopedia",
	"Prog Finished",
	"A Year in Thrills",
}

func Load() *Config {
	scanDirs := os.Getenv("SCAN_DIRS")
	return &Config{
		DatabasePath:    getEnv("DATABASE_PATH", "./reader.db"),
		ScanDirectories: splitDirs(scanDirs),
		LibraryName:     getEnv("LIBRARY_NAME", "2000 AD"),
		Host:            getEnv("HOST", ":8081"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ScanOnStartup:   getEnv("SCAN_ON_STARTUP", "false") == "true",
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func splitDirs(dirs string) []string {
	if dirs == "" {
		return []string{}
	}
	parts := strings.Split(dirs, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
