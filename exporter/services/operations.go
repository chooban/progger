package services

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2/data/binding"
	downloadApi "github.com/chooban/progger/download"
	"github.com/chooban/progger/exporter/api"
)

// ScanOperation represents an ongoing or completed scan operation
type ScanOperation struct {
	IsRunning binding.Bool
	Stories   binding.UntypedList
	Error     binding.String
	cancel    context.CancelFunc
}

// Cancel stops the scan operation if it's still running
func (s *ScanOperation) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}

// DownloadListOperation represents fetching the list of available issues from Rebellion
type DownloadListOperation struct {
	IsRunning      binding.Bool
	AvailableProgs binding.UntypedList
	Error          binding.String
	cancel         context.CancelFunc
}

// Cancel stops the download list operation if it's still running
func (d *DownloadListOperation) Cancel() {
	if d.cancel != nil {
		d.cancel()
	}
}

// DownloadOperation represents downloading one or more issues
type DownloadOperation struct {
	IsRunning binding.Bool
	Error     binding.String
	cancel    context.CancelFunc
}

// Cancel stops the download operation if it's still running
func (d *DownloadOperation) Cancel() {
	if d.cancel != nil {
		d.cancel()
	}
}

// NewScanOperation creates a new scan operation with initialized bindings
func NewScanOperation() *ScanOperation {
	return &ScanOperation{
		IsRunning: binding.NewBool(),
		Stories:   binding.NewUntypedList(),
		Error:     binding.NewString(),
	}
}

// NewDownloadListOperation creates a new download list operation with initialized bindings
func NewDownloadListOperation() *DownloadListOperation {
	return &DownloadListOperation{
		IsRunning:      binding.NewBool(),
		AvailableProgs: binding.NewUntypedList(),
		Error:          binding.NewString(),
	}
}

// NewDownloadOperation creates a new download operation with initialized bindings
func NewDownloadOperation() *DownloadOperation {
	return &DownloadOperation{
		IsRunning: binding.NewBool(),
		Error:     binding.NewString(),
	}
}

// BuildIssueList converts issues to untyped list and checks which are already downloaded
func BuildIssueList(issues []api.Downloadable, progSourceDir, megSourceDir string) []interface{} {
	if len(issues) == 0 {
		return make([]interface{}, 0)
	}

	// Sort by issue number descending
	sort.Slice(issues, func(a, b int) bool {
		return issues[a].Comic.IssueNumber > issues[b].Comic.IssueNumber
	})

	untypedIssues := make([]interface{}, len(issues))
	for i, v := range issues {
		targetDir := progSourceDir
		if v.Comic.Publication == "Megazine" {
			targetDir = megSourceDir
		}

		// Check if already downloaded
		filename := v.Comic.Filename(downloadApi.Pdf)
		if _, err := os.Stat(filepath.Join(targetDir, filename)); err == nil {
			v.Downloaded = true
		}
		untypedIssues[i] = v
	}

	return untypedIssues
}
