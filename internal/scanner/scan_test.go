package scanner

type EpisodeWriter struct {
	Title string
	Name  string
}

//func TestScanAndSave(t *testing.T) {
//	testing_init.IntegrationTest(t)
//	appEnv := createAppEnv()
//	appEnv.Db = db.Init("file:test.db?cache=shared&mode=memory")
//	appEnv.Pdf = pdfium.NewPdfiumReader(appEnv.Log)
//
//	t.Run("Scan directory", func(t *testing.T) {
//		dataDir := strings.Join([]string{"test", "testdata", "firstscan"}, string(os.PathSeparator))
//		secondDataDir := strings.Join([]string{"test", "testdata", "secondscan"}, string(os.PathSeparator))
//
//		issues := ScanDir(appEnv, dataDir, -1)
//		assert.Len(t, issues, 2, "Wrong number of issues found")
//
//		db.SaveIssues(appEnv.Db, issues)
//
//		var publications []db.Publication
//		var foundIssues []db.Issue
//		var episodes []db.Episode
//		var writers, artists, colourists, letterers []EpisodeWriter
//
//		appEnv.Db.Find(&publications)
//		appEnv.Db.Find(&foundIssues)
//		appEnv.Db.Find(&episodes)
//		appEnv.Db.Raw("SELECT title, name FROM episodes e, creators c, episode_writers ew where e.id = ew.episode_id and c.id = ew.creator_id").Scan(&writers)
//		appEnv.Db.Raw("SELECT title, name FROM episodes e, creators c, episode_artists ew where e.id = ew.episode_id and c.id = ew.creator_id").Scan(&artists)
//		appEnv.Db.Raw("SELECT title, name FROM episodes e, creators c, episode_colourists ew where e.id = ew.episode_id and c.id = ew.creator_id").Scan(&colourists)
//		appEnv.Db.Raw("SELECT title, name FROM episodes e, creators c, episode_letterers ew where e.id = ew.episode_id and c.id = ew.creator_id").Scan(&letterers)
//
//		assert.Equal(t, 1, len(publications), "Expected %d publications, found %d", 1, len(publications))
//		assert.Equal(t, 2, len(foundIssues), "Expected %d issues, found %d", 2, len(foundIssues))
//		assert.Equal(t, 14, len(episodes), "Expected %d episodes, found %d", 14, len(episodes))
//		assert.Equal(t, 14, len(writers), "Expected %d writers, found %d", 14, len(writers))
//		assert.Equal(t, 14, len(artists), "Expected %d artists, found %d", 14, len(artists))
//		assert.Equal(t, 7, len(colourists), "Expected %d colourists, found %d", 7, len(colourists))
//		assert.Equal(t, 14, len(letterers), "Expected %d letterers, found %d", 14, len(letterers))
//
//		// Do a second scan to ensure we don't duplicate any data
//		issues = ScanDir(appEnv, secondDataDir, -1)
//		assert.Len(t, issues, 3, "Wrong number of issues found")
//
//		db.SaveIssues(appEnv.Db, issues)
//
//		appEnv.Db.Find(&publications)
//		appEnv.Db.Find(&foundIssues)
//		appEnv.Db.Find(&episodes)
//		appEnv.Db.Raw("SELECT title, name FROM episodes e, creators c, episode_writers ew where e.id = ew.episode_id and c.id = ew.creator_id").Scan(&writers)
//
//		assert.Equal(t, 1, len(publications), "Expected %d publications, found %d", 1, len(publications))
//		assert.Equal(t, 3, len(foundIssues), "Expected %d issues, found %d", 3, len(foundIssues))
//		assert.Equal(t, 20, len(episodes), "Expected %d episodes, found %d", 20, len(episodes))
//		assert.Equal(t, 18, len(writers), "Expected %d writers, found %d", 18, len(writers))
//	})
//}
