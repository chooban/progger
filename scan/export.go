package scan

import (
	"context"
	"fmt"
	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal"
	"strings"
)

// Build exports a PDF of the pages passed to it
func Build(ctx context.Context, pages []api.ExportPage, artistsEdition bool, fileName string) error {
	if !strings.HasSuffix(fileName, "pdf") {
		return fmt.Errorf("file name must end with 'pdf'")
	}

	builder := internal.NewPdfBuilder()

	return builder.Build(pages, artistsEdition, fileName)
}
