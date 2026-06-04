package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chooban/progger/database"
	"github.com/chooban/progger/scan"
	scanApi "github.com/chooban/progger/scan/api"
	"github.com/go-logr/logr"
)

// Scanner defines the interface for scanning directories and files
type Scanner interface {
	Dir(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error)
	File(ctx context.Context, fileName string) (scanApi.Issue, error)
}

// RealScanner wraps the actual scan.Scanner to implement the Scanner interface
type RealScanner struct {
	scanner *scan.Scanner
}

// NewRealScanner creates a new RealScanner that uses the underlying scan.Scanner
func NewRealScanner(knownSeries, skipTitles []string) Scanner {
	return &RealScanner{
		scanner: scan.NewScanner(knownSeries, skipTitles),
	}
}

func (r *RealScanner) Dir(ctx context.Context, dir string, scanCount int) ([]scanApi.Issue, error) {
	// TODO: This needs to be async
	return r.scanner.Dir(ctx, dir, scanCount)
}

func (r *RealScanner) File(ctx context.Context, fileName string) (scanApi.Issue, error) {
	return r.scanner.File(ctx, fileName)
}

type ScanService struct {
	db         *database.DB
	cfg        *Config
	librarySer *database.LibraryRepo
	seriesSer  *database.SeriesRepo
	bookSer    *database.BookRepo
	coverSer   *database.CoverRepo
	scanner    Scanner
}

func NewScanService(db *database.DB, cfg *Config, librarySer *database.LibraryRepo, seriesSer *database.SeriesRepo, bookSer *database.BookRepo, coverSer *database.CoverRepo, scanner Scanner) *ScanService {
	return &ScanService{
		db:         db,
		cfg:        cfg,
		librarySer: librarySer,
		seriesSer:  seriesSer,
		bookSer:    bookSer,
		coverSer:   coverSer,
		scanner:    scanner,
	}
}

func (s *ScanService) Cfg() *Config {
	return s.cfg
}

func (s *ScanService) ScanDirectories(ctx context.Context) error {
	lib, err := s.librarySer.EnsureLibrary(ctx, s.cfg.LibraryName)
	if err != nil {
		return fmt.Errorf("failed to ensure library: %w", err)
	}

	for _, path := range s.cfg.ScanDirectories {
		if err := s.scanPath(ctx, lib.ID, path); err != nil {
			return fmt.Errorf("failed to scan %s: %w", path, err)
		}
	}

	return nil
}

func (s *ScanService) scanPath(ctx context.Context, libraryID int64, path string) error {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		println("Logger error: " + err.Error())
	}
	issues, err := s.scanner.Dir(ctx, path, 0)
	if err != nil {
		logger.Error(err, "failed to scan directory")
		return err
	}
	logger.Info("Finished scanning path", "issues", len(issues))

	seriesMap, err := s.upsertSeriesFromIssues(ctx, libraryID, issues)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		s.processIssue(ctx, libraryID, seriesMap, &issue)
	}

	return s.runPostScanSQL(ctx)
}

func (s *ScanService) upsertSeriesFromIssues(ctx context.Context, libraryID int64, issues []scanApi.Issue) (map[string]*database.Series, error) {
	seriesMap := make(map[string]*database.Series)
	for _, issue := range issues {
		for _, ep := range issue.Episodes {
			seriesName := ep.Series
			if seriesName == "" {
				continue
			}

			if _, ok := seriesMap[seriesName]; !ok {
				series := &database.Series{
					ID:        database.HashEntityID("series", seriesName),
					LibraryID: libraryID,
					Name:      seriesName,
				}
				if err := s.seriesSer.Upsert(ctx, series); err != nil {
					return nil, fmt.Errorf("failed to upsert series %s: %w", seriesName, err)
				}
				seriesMap[seriesName] = series
			}
		}
	}
	return seriesMap, nil
}

func (s *ScanService) processIssue(ctx context.Context, libraryID int64, seriesMap map[string]*database.Series, issue *scanApi.Issue) {
	logger := logr.FromContextOrDiscard(ctx)
	sublogger := logger.WithValues("issue_number", issue.IssueNumber)
	scanContext := logr.NewContext(ctx, sublogger)
	if len(issue.Episodes) == 0 {
		return
	}
	stories := toStories([]scanApi.Issue{*issue})
	for _, story := range stories {
		if err := s.createBook(scanContext, seriesMap[story.Series], issue, story); err != nil {
			logger.Error(err, "failed to create book", "story", story.Title)
		}
	}
	cover := issue.Cover
	if cover.Series != "" {
		if series, err := s.seriesSer.MaybeGetByName(scanContext, cover.Series); err == nil && series != nil {
			c := &database.Cover{
				ID:          database.HashEntityID("cover", series.Name, strconv.Itoa(issue.IssueNumber)),
				Text:        cover.Text,
				SeriesID:    series.ID,
				Artist:      cover.Artist,
				Filename:    cover.Filename,
				IssueNumber: issue.IssueNumber,
				Publication: issue.Publication,
			}
			if err := s.coverSer.Upsert(scanContext, c); err != nil {
				logger.Error(err, "failed to save cover")
			}
		}
	}
}

func (s *ScanService) runPostScanSQL(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
with book_counts as (
	select series_id, count(*) as c
	from books b
	group by series_id
	)
update series set book_count = ( select c from book_counts where book_counts.series_id = series.id )
`)

	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
with page_counts as (
	select book_id, sum((e.page_to - e.page_from) + 1) as c
	from episodes e
	group by 1
)
update books set page_count = (
	select c from page_counts where page_counts.book_id = books.id
);
`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
with first_issues as (
select distinct
	books.id,
	first_value(episodes.issue_number) OVER (partition by books.id ORDER BY issue_number) as issue_number
from books
join episodes on (episodes.book_id = books.id)
),
last_issues as (
	select distinct
		books.id,
		first_value(episodes.issue_number) OVER (partition by books.id ORDER BY issue_number DESC) as issue_number
	from books
	join episodes on (episodes.book_id = books.id)
)
update books
set first_issue = (select issue_number from first_issues where id = books.id), last_issue = (select issue_number from last_issues where id = books.id)
;`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
with book_order as (
	select 
		books.id, 
		row_number() OVER (PARTITION BY series_id ORDER BY first_issue) as row_number
	from books 
	order by id ASC
)
update books set number = ( select row_number from book_order where id = books.id limit 1 )
`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
with book_release_date as (
	select 
		books.id, 
		first_value(episodes.release_date) OVER (PARTITION BY books.id ORDER BY issue_number) as first_release_date
	from books 
	join episodes on (episodes.book_id = books.id)
)
update books set release_date = ( select first_release_date from book_release_date where id = books.id limit 1 )
`)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScanService) RefreshMetadata(ctx context.Context, libraryID int64) error {
	logger := logr.FromContextOrDiscard(ctx)

	books, _, err := s.bookSer.ListByLibraryID(ctx, libraryID, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list books: %w", err)
	}

	seen := make(map[string]bool)
	var files []string
	for _, book := range books {
		for _, ep := range book.Episodes {
			if !seen[ep.Filename] {
				seen[ep.Filename] = true
				files = append(files, ep.Filename)
			}
		}
	}

	seriesMap := make(map[string]*database.Series)
	for i, filename := range files {
		issue, err := s.scanner.File(ctx, filename)
		if err != nil {
			logger.Error(err, "failed to scan file", "filename", filename)
			continue
		}

		for _, ep := range issue.Episodes {
			seriesName := ep.Series
			if seriesName == "" {
				continue
			}
			if _, ok := seriesMap[seriesName]; !ok {
				series := &database.Series{
					ID:        database.HashEntityID("series", seriesName),
					LibraryID: libraryID,
					Name:      seriesName,
				}
				if err := s.seriesSer.Upsert(ctx, series); err != nil {
					return fmt.Errorf("failed to upsert series %s: %w", seriesName, err)
				}
				seriesMap[seriesName] = series
			}
		}

		s.processIssue(ctx, libraryID, seriesMap, &issue)
		logger.Info("refreshed metadata", "file", filename, "progress", fmt.Sprintf("%d/%d", i+1, len(files)))
	}

	return s.runPostScanSQL(ctx)
}

func toStories(issues []scanApi.Issue) []*Story {
	storyMap := make(map[string]*Story)

	for _, issue := range issues {
		for _, ep := range issue.Episodes {
			key := ep.Series + " - " + ep.Title
			if story, ok := storyMap[key]; ok {
				story.Episodes = append(story.Episodes, Episode{
					Episode:     ep,
					Filename:    issue.Filename,
					IssueNumber: issue.IssueNumber,
				})
			} else {
				story = &Story{
					Title:      ep.Title,
					Series:     ep.Series,
					FirstIssue: issue.IssueNumber,
					LastIssue:  issue.IssueNumber,
					Episodes:   []Episode{{Episode: ep}},
				}
				story.Episodes[0].Filename = issue.Filename
				story.Episodes[0].IssueNumber = issue.IssueNumber
				storyMap[key] = story
			}
		}
	}

	stories := make([]*Story, 0, len(storyMap))
	for _, s := range storyMap {
		stories = append(stories, s)
	}
	return stories
}

func (s *ScanService) createBook(ctx context.Context, series *database.Series, issue *scanApi.Issue, story *Story) error {
	logger := logr.FromContextOrDiscard(ctx)
	book := &database.Book{
		ID:          database.HashEntityID("book", series.Name, story.Title),
		SeriesID:    series.ID,
		Name:        story.Title,
		Status:      "READY",
		PageCount:   0,
		FirstIssue:  0,
		LastIssue:   0,
		Publication: issue.Publication,
	}

	episodes := make([]*database.Episode, len(story.Episodes))
	for i, ep := range story.Episodes {
		coverDate := issue.CoverDate
		episodes[i] = &database.Episode{
			ID:          database.HashEntityID("episode", series.Name, story.Title, strconv.Itoa(ep.Episode.Part)),
			Filename:    ep.Filename,
			IssueNumber: ep.IssueNumber,
			Title:       ep.Episode.Title,
			Part:        ep.Episode.Part,
			PageFrom:    ep.Episode.FirstPage,
			PageTo:      ep.Episode.LastPage,
			ReleaseDate: &coverDate,
		}
	}

	logger.V(1).Info("going to upsert book", "book", book.Name)
	if err := s.bookSer.Upsert(ctx, book); err != nil {
		return fmt.Errorf("failed to upsert book: %w", err)
	}

	for _, episode := range episodes {
		episode.BookID = book.ID
	}

	if err := s.bookSer.UpsertEpisodes(ctx, episodes); err != nil {
		return fmt.Errorf("failed to insert episodes: %w", err)
	}

	return nil
}
