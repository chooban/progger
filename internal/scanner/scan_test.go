package scanner

import (
	"github.com/chooban/progdl-go/internal/db"
	"github.com/chooban/progdl-go/internal/pdfium"
	_ "github.com/chooban/progdl-go/testing_init"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestScanAndSave(t *testing.T) {
	integrationTest(t)
	appEnv := createAppEnv()
	appEnv.Db = db.Init("file:test.db?cache=shared&mode=memory")
	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)

	t.Run("Scan directory", func(t *testing.T) {
		dataDir := strings.Join([]string{"test", "testdata", "firstscan"}, string(os.PathSeparator))
		secondDataDir := strings.Join([]string{"test", "testdata", "secondscan"}, string(os.PathSeparator))

		issues := ScanDir(appEnv, dataDir, -1)
		assert.Len(t, issues, 2, "Wrong number of issues found")

		db.SaveIssues(appEnv.Db, issues)

		var publications []db.Publication
		var foundIssues []db.Issue
		var episodes []db.Episode

		appEnv.Db.Find(&publications)
		appEnv.Db.Find(&foundIssues)
		appEnv.Db.Find(&episodes)

		assert.Equal(t, 1, len(publications), "Expected %d publications, found %d", 1, len(publications))
		assert.Equal(t, 2, len(foundIssues), "Expected %d issues, found %d", 2, len(foundIssues))
		assert.Equal(t, 14, len(episodes), "Expected %d episodes, found %d", 14, len(episodes))

		// Do a second scan to ensure we don't duplicate any data
		issues = ScanDir(appEnv, secondDataDir, -1)
		assert.Len(t, issues, 3, "Wrong number of issues found")

		db.SaveIssues(appEnv.Db, issues)

		appEnv.Db.Find(&publications)
		appEnv.Db.Find(&foundIssues)
		appEnv.Db.Find(&episodes)

		assert.Equal(t, 1, len(publications), "Expected %d publications, found %d", 1, len(publications))
		assert.Equal(t, 3, len(foundIssues), "Expected %d issues, found %d", 3, len(foundIssues))
		assert.Equal(t, 20, len(episodes), "Expected %d episodes, found %d", 20, len(episodes))
	})
}

func integrationTest(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests, set environment variable INTEGRATION")
	}
}
