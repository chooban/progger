package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/chooban/progger/scan/testing_init"
	"github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/klippa-app/go-pdfium/requests"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func IntegrationTest(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests, set environment variable INTEGRATION")
	}
}

func TestPdfiumReader_Credits(t *testing.T) {
	IntegrationTest(t)
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var log = zerologr.New(&logger)

	pdfium := NewPdfiumReader(log)
	dataDir := strings.Join([]string{"test", "testdata", "creators"}, string(os.PathSeparator))

	testCases := []struct {
		name        string
		filename    string
		page        int
		wantCredits string
	}{
		{
			name:        "Dredd - 1999",
			filename:    "2000AD 1999 (1977).pdf",
			page:        3,
			wantCredits: "script t.c. eglington colours chris blythe art paul marshall letters annie parkhouse",
		},
		{
			name:        "Jaegir - 1999",
			filename:    "2000AD 1999 (1977).pdf",
			page:        9,
			wantCredits: "art simon coleby script gordon rennie letters simon bowland colours len o’grady",
		},
		{
			name:        "Scarlet Traces - 1999",
			filename:    "2000AD 1999 (1977).pdf",
			page:        14,
			wantCredits: "script ian edginton letters annie parkhouse art d’israeli",
		},
		{
			name:        "Outlier - 1999",
			filename:    "2000AD 1999 (1977).pdf",
			page:        19,
			wantCredits: "script t.c. eglington letters ellie de ville art karl richardson",
		},
		{
			name:        "Anderson - 1999",
			filename:    "2000AD 1999 (1977).pdf",
			page:        25,
			wantCredits: "script emma beeby colours richard elson art ben willsher letters ellie de ville",
		},
		{
			name:        "Dredd - 2300",
			filename:    "2000AD 2300 (1977).pdf",
			page:        3,
			wantCredits: "script ken niemand art henry flint letters annie parkhouse",
		},
		{
			name:        "Rogue Trooper - 2300",
			filename:    "2000AD 2300 (1977).pdf",
			page:        11,
			wantCredits: "script mike carroll colours yel zamor art gary erskine letters simon bowland",
		},
		{
			name:        "Survival Geeks - 2300",
			filename:    "2000AD 2300 (1977).pdf",
			page:        17,
			wantCredits: "script emma beeby colours gary caldwell art neil googe letters jim campbell",
		},
		{
			name:        "Sinister Dexter - 2300",
			filename:    "2000AD 2300 (1977).pdf",
			page:        26,
			wantCredits: "script dan abnett letters annie parkhouse art russell m. olson",
		},
		{
			name:        "Proteus Vex - 2272",
			filename:    "2000AD 2272 (1977).pdf",
			page:        10,
			wantCredits: "script mike carroll colours jim boswell art jake lynch letters simon bowland",
		},
		{
			name:        "Azimuth - 2337",
			filename:    "2000AD 2337 (1977).pdf",
			page:        20,
			wantCredits: "script dan abnett colours matt soffe art tazio bettin letters jim campbell",
		},
		{
			name:        "Slaine - 2215",
			filename:    "2000AD 2215 (1977).pdf",
			page:        14,
			wantCredits: "script pat mills art leonardo manco letters annie parkhouse",
		},
		{
			name:        "Dredd - 2317",
			filename:    "2000AD 2317 (1977).pdf",
			page:        3,
			wantCredits: "script rob williams & arthur wyatt colours dylan teague art paul marshall letters annie parkhouse",
		},
		{
			name:        "Anderson - 2183",
			filename:    "2000AD 2183 (1977).pdf",
			page:        23,
			wantCredits: "art paul davidson script cavan scott colours len o’grady letters simon bowland",
		},
		{
			name:        "Hershey - 2348",
			filename:    "2000AD 2348 (1977).pdf",
			page:        25,
			wantCredits: "script rob williams art simon fraser letters simon bowland",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			credits, err := pdfium.Credits(
				strings.Join([]string{dataDir, tc.filename}, string(os.PathSeparator)),
				tc.page, tc.page+1,
			)
			assert.Nil(t, err, "Error should be nil: %s", err)
			assert.Equal(t, tc.wantCredits, credits)
		})
	}
}

func testLogger() logr.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return zerologr.New(&logger)
}

func creatorsDataDir() string {
	return strings.Join([]string{"test", "testdata", "creators"}, string(os.PathSeparator))
}

func createSyntheticPDF(t *testing.T, path string, numPages int) {
	t.Helper()

	doc, err := Instance.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
	assert.Nil(t, err)

	for i := 0; i < numPages; i++ {
		page, err := Instance.FPDFPage_New(&requests.FPDFPage_New{
			Document:  doc.Document,
			PageIndex: i,
			Width:     595.0,
			Height:    842.0,
		})
		assert.Nil(t, err)

		_, err = Instance.FPDFPage_GenerateContent(&requests.FPDFPage_GenerateContent{
			Page: requests.Page{ByReference: &page.Page},
		})
		assert.Nil(t, err)

		_, err = Instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: page.Page})
		assert.Nil(t, err)
	}

	_, err = Instance.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{
		Flags:    requests.SaveFlagNoIncremental,
		Document: doc.Document,
		FilePath: &path,
	})
	assert.Nil(t, err)

	Instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
}

func TestPdfiumReader_Bookmarks(t *testing.T) {
	IntegrationTest(t)
	log := testLogger()
	reader := NewPdfiumReader(log)
	dataDir := creatorsDataDir()

	type wantBm struct {
		Title    string
		PageFrom int
		PageThru int
	}

	tests := []struct {
		name      string
		filename  string
		wantCount int
		bookmarks []wantBm
	}{
		{
			name:      "1999",
			filename:  "2000AD 1999 (1977).pdf",
			wantCount: 7,
			bookmarks: []wantBm{
				{"Cover", 1, 1},
				{"Nerve Centre", 2, 2},
				{"Judge Dredd: Well Gel", 3, 8},
				{"Jaegir: Warchild - Part 4", 9, 13},
				{"Scarlet Traces: Cold War - Part 12", 14, 18},
				{"Outlier: Survivor Guilt - Part 10 ", 19, 24},
				{"Anderson Psi Division: The Candidate - Part 7", 25, 32},
			},
		},
		{
			name:      "2300",
			filename:  "2000AD 2300 (1977).pdf",
			wantCount: 11,
			bookmarks: []wantBm{
				{"Cover", 1, 1},
				{"Tharg's Nerve Centre", 2, 2},
				{"Judge Dredd: Judgement Days - Prologue ", 3, 10},
				{"Rogue Trooper - Mortal Remains", 11, 16},
				{"Survival Geeks- House of The Dead", 17, 20},
				{"The Meat Arena", 21, 25},
				{"Sinister Dexter - Zed Zone", 26, 31},
				{"Ampney Crucis Investigates... - Setting Sons", 32, 36},
				{"Robo-Hunter: Z-INF", 37, 40},
				{"Strontium Dog: In The (Dead) Doghouse", 41, 45},
				{"Judge Dredd: Judgement Days - Epilogue", 46, 52},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(dataDir, tt.filename)
			details, err := reader.Bookmarks(fullPath)
			assert.Nil(t, err)
			assert.Len(t, details, tt.wantCount)

			for i, want := range tt.bookmarks {
				if i >= len(details) {
					break
				}
				assert.Equal(t, want.Title, details[i].Bookmark.Title, "bookmark %d title", i)
				assert.Equal(t, want.PageFrom, details[i].Bookmark.PageFrom, "bookmark %d PageFrom", i)
				assert.Equal(t, want.PageThru, details[i].Bookmark.PageThru, "bookmark %d PageThru", i)
			}
		})
	}
}

func TestPdfBuilder_BuildPageAsPDF(t *testing.T) {
	IntegrationTest(t)

	t.Run("standard", func(t *testing.T) {
		assertBuildPageAsPDF(t, false)
	})
	t.Run("artistsEdition", func(t *testing.T) {
		assertBuildPageAsPDF(t, true)
	})
}

func assertBuildPageAsPDF(t *testing.T, artistsEdition bool) {
	t.Helper()
	dataDir := creatorsDataDir()
	pdfPath := filepath.Join(dataDir, "2000AD 1999 (1977).pdf")

	builder := NewPdfBuilder()
	page := api.ExportPage{
		Filename: pdfPath,
		PageFrom: 1,
		PageTo:   2,
	}

	result, err := builder.BuildPageAsPDF(page, artistsEdition)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(*result), 0, "output PDF bytes must not be empty")
}

func TestPdfBuilder_AddBookmarks(t *testing.T) {
	IntegrationTest(t)

	srcPath := filepath.Join(t.TempDir(), "source.pdf")
	outPath := filepath.Join(t.TempDir(), "output.pdf")
	createSyntheticPDF(t, srcPath, 3)

	builder := NewPdfBuilder()
	builder.OpenDestination()

	builder.CopyPages(&srcPath, 1, 3, 0)
	assert.Nil(t, builder.BuildError)

	builder.Save(outPath)
	assert.Nil(t, builder.BuildError)

	bookmarks := []pdfcpu.Bookmark{
		{Title: "Chapter One", PageFrom: 1, PageThru: 2},
		{Title: "Chapter Two", PageFrom: 3, PageThru: 0},
	}
	builder.AddBookmarks(bookmarks)
	assert.Nil(t, builder.BuildError)

	f, err := os.Open(outPath)
	assert.Nil(t, err)
	defer f.Close()

	readBack, err := pdfApi.Bookmarks(f, model.NewDefaultConfiguration())
	assert.Nil(t, err)
	assert.Len(t, readBack, 2)
	assert.Equal(t, "Chapter One", readBack[0].Title)
	assert.Equal(t, 1, readBack[0].PageFrom)
	assert.Equal(t, 2, readBack[0].PageThru)
	assert.Equal(t, "Chapter Two", readBack[1].Title)
	assert.Equal(t, 3, readBack[1].PageFrom)
	assert.Equal(t, 0, readBack[1].PageThru)
}
