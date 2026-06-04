package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"strconv"
	"sync"

	"github.com/chooban/progger/database"
	"github.com/chooban/progger/scan"
	scanApi "github.com/chooban/progger/scan/api"
	"github.com/disintegration/imaging"
	"github.com/go-logr/logr"
)

type thumbnailBuilderFunc func(context.Context, scanApi.ExportPage) (*image.RGBA, error)

type ThumbnailService struct {
	pageService         *PageService
	seriesCoverBuilder  thumbnailBuilderFunc
	cache               sync.Map
}

func NewThumbnailService(pageService *PageService) *ThumbnailService {
	return &ThumbnailService{
		pageService:        pageService,
		seriesCoverBuilder: scan.BuildPageAsImage,
	}
}

// resizeImage resizes an image to fit within maxDim on the largest dimension,
// maintaining aspect ratio. Uses Lanczos filter for high-quality downsampling.
func resizeImage(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling factor based on largest dimension
	var scale float64
	if width > height {
		scale = float64(maxDim) / float64(width)
	} else {
		scale = float64(maxDim) / float64(height)
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Ensure we have at least 1 pixel dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Use Lanczos filter for high-quality downsampling
	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

func (s *ThumbnailService) GetSeriesThumbnail(ctx context.Context, cover *database.Cover, seriesName string) ([]byte, error) {
	cacheKey := "series-" + seriesName

	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.([]byte), nil
	}

	exportPage := scanApi.ExportPage{
		Filename:       cover.Filename,
		PageFrom:       1,
		PageTo:         2,
		ArtistsEdition: true,
	}

	rgba, err := s.seriesCoverBuilder(ctx, exportPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build page: %w", err)
	}

	resized := resizeImage(rgba, 300)
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: 95}); err != nil {
		return nil, fmt.Errorf("failed to encode jpeg: %w", err)
	}

	jpegData := buf.Bytes()

	s.cache.Store(cacheKey, jpegData)

	return jpegData, nil
}

func (s *ThumbnailService) GetCoverThumbnail(ctx context.Context, cover *database.Cover) ([]byte, error) {
	var cacheKey string

	if cover != nil {
		cacheKey = "cover-" + strconv.Itoa(int(cover.ID))
	} else {
		return []byte{}, errors.New("cover is nil")
	}

	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.([]byte), nil
	}

	var thumbnailFile string
	if cover.Filename != "" {
		thumbnailFile = cover.Filename
	} else {
		return []byte("no thumbnail available"), nil
	}

	exportPage := scanApi.ExportPage{
		Filename:       thumbnailFile,
		PageFrom:       1,
		PageTo:         2,
		ArtistsEdition: true,
	}

	rgba, err := s.pageService.coverBuilder(ctx, exportPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build page: %w", err)
	}

	// Resize to 300px on largest dimension
	resized := resizeImage(rgba, 300)

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("failed to encode jpeg: %w", err)
	}

	jpegData := buf.Bytes()

	s.cache.Store(cacheKey, jpegData)

	return jpegData, nil
}

// GetPageThumbnail retrieves a page as an image, resizes it to 300px on the largest dimension,
// and encodes it as JPEG.
func (s *ThumbnailService) GetPageThumbnail(ctx context.Context, bookID int64, pageNum int) ([]byte, error) {
	logger := logr.FromContextOrDiscard(ctx)
	cacheKey := fmt.Sprintf("page-thumbnail-%d-%d", bookID, pageNum)

	// Check cache first
	if cached, ok := s.cache.Load(cacheKey); ok {
		logger.V(1).Info("thumbnail cache hit")
		return cached.([]byte), nil
	}

	// Get the page as an image using the book's episodes
	book, err := s.pageService.bookSer.GetByID(ctx, bookID)
	if err != nil {
		logger.Info("failed to get book for thumbnail")
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	episodes := book.Episodes
	ep, epPage, err := s.pageService.FindEpisodeForPage(episodes, pageNum)
	if err != nil {
		logger.Info("failed to get episode for thumbnail")
		return nil, fmt.Errorf("failed to find episode for page: %w", err)
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

	// Build page as image
	logger.V(0).Info("making thumbnail request", "page", exportPage)
	rgba, err := s.pageService.coverBuilder(ctx, exportPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build page image: %w", err)
	}

	// Resize to 300px on largest dimension
	resized := resizeImage(rgba, 300)

	// Encode to JPEG with quality 85
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("failed to encode jpeg: %w", err)
	}

	jpegData := buf.Bytes()

	// Cache the result
	s.cache.Store(cacheKey, jpegData)

	return jpegData, nil
}
