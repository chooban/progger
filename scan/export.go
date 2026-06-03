package scan

import (
	"context"
	"fmt"
	"image"
	"strings"

	"github.com/chooban/progger/scan/api"
	"github.com/chooban/progger/scan/internal"
)

// Build exports a PDF of the pages passed to it
func Build(ctx context.Context, pages []api.ExportPage, fileName string) error {
	if !strings.HasSuffix(fileName, "pdf") {
		return fmt.Errorf("file name must end with 'pdf'")
	}

	builder := internal.NewPdfBuilder()

	return builder.Build(pages, fileName)
}

func BuildPageAsImage(ctx context.Context, page api.ExportPage) (*image.RGBA, error) {
	builder := internal.NewPdfBuilder()

	return builder.BuildPage(page)
}

func BuildPageAsPDF(ctx context.Context, page api.ExportPage) (*[]byte, error) {
	builder := internal.NewPdfBuilder()

	return builder.BuildPageAsPDF(page)
}
