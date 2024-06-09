package services

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	api2 "github.com/chooban/progger/exporter/api"
	"github.com/chooban/progger/scan"
	"github.com/chooban/progger/scan/api"
	"path/filepath"
	"slices"
)

type Exporter struct {
	ctx context.Context
}

func (e *Exporter) Export(stories []*api2.Story, sourceDir, exportDir, filename string) error {
	toExport := make([]api.ExportPage, 0)
	for _, story := range stories {
		if story.ToExport {
			for _, e := range story.Episodes {
				toExport = append(toExport, api.ExportPage{
					Filename:    filepath.Join(sourceDir, e.Filename),
					PageFrom:    e.FirstPage,
					PageTo:      e.LastPage,
					IssueNumber: e.IssueNumber,
					Title:       fmt.Sprintf("%s - Part %d", e.Title, e.Part),
				})
			}
		}
	}
	if len(toExport) == 0 {
		return errors.New("no stories to export")
	}
	// Sort by issue number. We sometimes have issues being wrongly grouped, but surely we never want anything
	// other than issue order?
	slices.SortFunc(toExport, func(i, j api.ExportPage) int {
		return cmp.Compare(i.IssueNumber, j.IssueNumber)
	})

	// Do the export
	err := scan.Build(e.ctx, toExport, filepath.Join(exportDir, filename))
	if err != nil {
		return err
	}

	return nil
}

func NewExporter(ctx context.Context) *Exporter {
	return &Exporter{
		ctx,
	}
}
