package app

import (
	"context"

	"fyne.io/fyne/v2/data/binding"
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
	isRunning, _ := s.IsRunning.Get()
	if isRunning && s.cancel != nil {
		s.cancel()
	}
}

// SetCancel sets the cancel function for this operation
func (s *ScanOperation) SetCancel(cancel context.CancelFunc) {
	s.cancel = cancel
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
	isRunning, _ := d.IsRunning.Get()
	if isRunning && d.cancel != nil {
		d.cancel()
	}
}

// SetCancel sets the cancel function for this operation
func (d *DownloadListOperation) SetCancel(cancel context.CancelFunc) {
	d.cancel = cancel
}

// DownloadOperation represents downloading one or more issues
type DownloadOperation struct {
	IsRunning binding.Bool
	Error     binding.String
	cancel    context.CancelFunc
}

// Cancel stops the download operation if it's still running
func (d *DownloadOperation) Cancel() {
	isRunning, _ := d.IsRunning.Get()
	if isRunning && d.cancel != nil {
		d.cancel()
	}
}

// SetCancel sets the cancel function for this operation
func (d *DownloadOperation) SetCancel(cancel context.CancelFunc) {
	d.cancel = cancel
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
