package server

import (
	"context"

	scanApi "github.com/chooban/progger/scan/api"
)

// MockScanner is a mock implementation of Scanner for testing
type MockScanner struct {
	DirFunc  func(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error)
	FileFunc func(ctx context.Context, fileName string) (scanApi.Issue, error)
}

func (m *MockScanner) Dir(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
	if m.DirFunc != nil {
		return m.DirFunc(ctx, dir, scanCount)
	}
	return []scanApi.Issue{}, nil
}

func (m *MockScanner) File(ctx context.Context, fileName string) (scanApi.Issue, error) {
	if m.FileFunc != nil {
		return m.FileFunc(ctx, fileName)
	}
	return scanApi.Issue{}, nil
}
