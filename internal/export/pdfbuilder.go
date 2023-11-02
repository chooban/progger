package export

import (
	"fmt"
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/env"
	"github.com/klippa-app/go-pdfium/requests"
)

// BuildPdf will export a PDF of the provided series and optional
// episodes.
// The parameters of seriesTitle and episodeTitle should be used to
// query the database via appEnv.Db to retrieve all applicable episodes,
// ordered by issue number.
func BuildPdf(appEnv env.AppEnv, seriesTitle string, episodeTitle string) {
	var episodes []db.Episode

	appEnv.Db.Preload("Issue").Table("episodes e").
		Joins("join series s on s.id = e.series_id").
		Joins("join issues i on e.issue_id = i.id").
		Where("s.title = ? and e.issue_id > 0", seriesTitle).
		Order("e.title, part ASC").
		Find(&episodes)

	destination, err := appEnv.Pdfium.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
	if err != nil {
		appEnv.Log.Err(err).Msg("Could not create new document")
	}

	pageCount := 0
	for _, v := range episodes {
		filename := fmt.Sprintf("/Users/ross/Documents/2000AD/%s", v.Issue.Filename)
		source, err := appEnv.Pdfium.FPDF_LoadDocument(&requests.FPDF_LoadDocument{
			Path: &filename,
		})
		if err != nil {
			appEnv.Log.Err(err).Msg(fmt.Sprintf("Could not open PDF: %s", filename))
			continue
		}

		pageRange := fmt.Sprintf("%d-%d", v.PageFrom, v.PageThru)
		_, err = appEnv.Pdfium.FPDF_ImportPages(&requests.FPDF_ImportPages{
			Source:      source.Document,
			Destination: destination.Document,
			PageRange:   &pageRange,
			Index:       pageCount,
		})
		if err != nil {
			appEnv.Log.Err(err).Msg("Could not import pages")
			continue
		}

		pageCount = (v.PageThru - v.PageFrom) + 1
	}

	var output = "myfile.pdf"
	if saveAsCopy, err := appEnv.Pdfium.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    1,
		Document: destination.Document,
		FilePath: &output,
	}); err != nil {
		appEnv.Log.Err(err).Msg("Could not save document")
	} else {
		appEnv.Log.Info().Msg(fmt.Sprintf("File saved to %s", *saveAsCopy.FilePath))
	}
}
