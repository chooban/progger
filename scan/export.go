package scan

import (
	"context"
	"fmt"
	"image"
	"strings"

	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal"
)

type ExportFormat int

const (
	PDF   ExportFormat = 1
	Image ExportFormat = 2
)

// Build exports a PDF of the pages passed to it
func Build(ctx context.Context, pages []api.ExportPage, artistsEdition bool, fileName string) error {
	if !strings.HasSuffix(fileName, "pdf") {
		return fmt.Errorf("file name must end with 'pdf'")
	}

	builder := internal.NewPdfBuilder()

	return builder.Build(pages, artistsEdition, fileName)
}

func BuildPageAsImage(ctx context.Context, page api.ExportPage, artistsEdition bool) (*image.RGBA, error) {
	builder := internal.NewPdfBuilder()

	return builder.BuildPage(page, artistsEdition)
}

func BuildPageAsPDF(ctx context.Context, page api.ExportPage, artistsEdition bool) (*[]byte, error) {
	builder := internal.NewPdfBuilder()

	return builder.BuildPageAsPDF(page, artistsEdition)
}
