package scan

import (
	"context"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal/pdfium"
	"github.com/go-logr/logr"
	"strings"
)

func Build(ctx context.Context, pages []api.ExportPage, fileName string) error {
	logger := logr.FromContextOrDiscard(ctx)
	if !strings.HasSuffix(fileName, "pdf") {
		return fmt.Errorf("file name must end with 'pdf'")
	}
	p := pdfium.NewPdfiumReader(logger)

	p.Build(pages, fileName)

	return nil
}
