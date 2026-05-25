package services

import (
	"context"
	"fmt"
	"image"
	"sync"

	"github.com/chooban/progger/reader/models"
	scanApi "github.com/chooban/progger/scan/api"
)

type pageBuilderFunc func(context.Context, scanApi.ExportPage, bool) (*[]byte, error)
type coverBuilderFunc func(context.Context, scanApi.ExportPage, bool) (*image.RGBA, error)

type PageService struct {
	bookSer      *BookService
	pageCache    *pageCache
	bookCache    *bookCache
	pageBuilder  pageBuilderFunc
	coverBuilder coverBuilderFunc
}

type pageCache struct {
	sync.RWMutex
	data map[string][]byte
}

type bookCache struct {
	sync.RWMutex
	data map[int64][]*models.Episode
}

func NewPageService(bookSer *BookService, _ any, pageBuilder pageBuilderFunc, coverBuilder coverBuilderFunc) *PageService {
	return &PageService{
		bookSer:      bookSer,
		pageCache:    &pageCache{data: make(map[string][]byte)},
		bookCache:    &bookCache{data: make(map[int64][]*models.Episode)},
		pageBuilder:  pageBuilder,
		coverBuilder: coverBuilder,
	}
}

func (s *PageService) GetPages(ctx context.Context, bookID int64) ([]models.Page, error) {
	book, err := s.bookSer.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	if len(book.Episodes) == 0 {
		return []models.Page{}, nil
	}

	pagesCount := 0
	for _, ep := range book.Episodes {
		pagesCount += ep.PageTo - ep.PageFrom + 1
	}

	pages := make([]models.Page, pagesCount)
	pageNum := 1
	for i := 0; i < pagesCount; i++ {
		pages[i] = models.BuildPage(bookID, pageNum)
		pageNum++
	}

	return pages, nil
}

func (s *PageService) GetPage(ctx context.Context, bookID int64, pageNum int) ([]byte, error) {
	book, err := s.bookSer.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}
	episodes := book.Episodes
	ep, epPage, err := s.FindEpisodeForPage(episodes, pageNum)
	if err != nil {
		return nil, fmt.Errorf("failed to find episode: %w", err)
	}

	cacheKey := fmt.Sprintf("page-%d-%d-%d", bookID, pageNum, ep.ID)

	s.pageCache.RLock()
	if cached, ok := s.pageCache.data[cacheKey]; ok {
		s.pageCache.RUnlock()
		return cached, nil
	}
	s.pageCache.RUnlock()

	pageFrom := ep.PageFrom + epPage
	pageTo := pageFrom

	exportPage := scanApi.ExportPage{
		Filename:    ep.Filename,
		IssueNumber: ep.IssueNumber,
		Title:       ep.Title,
		PageFrom:    pageFrom,
		PageTo:      pageTo,
	}

	pdfData, err := s.pageBuilder(ctx, exportPage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build page: %w", err)
	}

	s.pageCache.Lock()
	s.pageCache.data[cacheKey] = *pdfData
	s.pageCache.Unlock()

	return *pdfData, nil
}

func (s *PageService) GetPageImage(ctx context.Context, bookID int64, pageNum int) (*image.RGBA, error) {
	book, err := s.bookSer.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}
	episodes := book.Episodes
	ep, epPage, err := s.FindEpisodeForPage(episodes, pageNum)
	if err != nil {
		return nil, fmt.Errorf("failed to find episode: %w", err)
	}

	pageFrom := ep.PageFrom + epPage
	pageTo := pageFrom

	exportPage := scanApi.ExportPage{
		Filename:    ep.Filename,
		IssueNumber: ep.IssueNumber,
		Title:       ep.Title,
		PageFrom:    pageFrom,
		PageTo:      pageTo,
	}

	rgba, err := s.coverBuilder(ctx, exportPage, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build page image: %w", err)
	}
	return rgba, nil
}

func (s *PageService) FindEpisodeForPage(episodes []*models.Episode, pageNum int) (*models.Episode, int, error) {
	if len(episodes) == 0 {
		return nil, 0, fmt.Errorf("no episodes")
	}

	runningPage := 1
	for _, ep := range episodes {
		pageCount := ep.PageTo - ep.PageFrom + 1
		if pageNum >= runningPage && pageNum < runningPage+pageCount {
			return ep, pageNum - runningPage, nil
		}
		runningPage += pageCount
	}

	return nil, 0, fmt.Errorf("page %d not found", pageNum)
}
