package pdfium

import (
	"github.com/chooban/progdl-go/testing_init"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPdfiumReader_Credits(t *testing.T) {
	testing_init.IntegrationTest(t)
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(writer)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	pdfium := NewPdfiumReader(&logger)
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
