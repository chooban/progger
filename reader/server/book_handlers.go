package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gangleri.io/pkg/humanbytes"
	"github.com/chooban/progger/reader/api"
	"github.com/chooban/progger/reader/models"
	"github.com/chooban/progger/reader/services"
	"github.com/chooban/progger/scan"
	scanApi "github.com/chooban/progger/scan/api"
	"github.com/gin-gonic/gin"
)

func bookToDto(b *models.Book, series *models.Series) api.BookDto {
	libraryID := ""
	if series != nil {
		libraryID = services.Int64ToStringID(series.LibraryID)
	}
	numberStr := strconv.Itoa(b.Number)

	episodes := b.Episodes
	summary := ""
	if len(episodes) > 0 {
		if len(episodes) == 1 {
			summary = fmt.Sprintf("Prog %d", episodes[0].IssueNumber)
		} else {
			summary = fmt.Sprintf("Progs %d - %d", episodes[0].IssueNumber, episodes[len(episodes)-1].IssueNumber)
		}
	}

	seriesName := ""
	if series != nil {
		seriesName = series.Name
	}
	sizeBytes := int64(b.PageCount * 768 * 1024)
	sizeStr, _ := humanbytes.Convert(int(sizeBytes), "MB") // 1MB
	createdAt := b.CreatedAt
	if createdAt == "" {
		createdAt = b.UpdatedAt
	}
	return api.BookDto{
		ID:               services.Int64ToStringID(b.ID),
		SeriesID:         services.Int64ToStringID(b.SeriesID),
		SeriesTitle:      seriesName,
		Name:             strconv.Itoa(b.Number) + " " + b.Name,
		Number:           int32(b.Number),
		SizeBytes:        sizeBytes,
		Size:             strconv.Itoa(int(sizeStr)) + "MB",
		URL:              b.URL(),
		Created:          createdAt,
		LastModified:     b.UpdatedAt,
		LibraryID:        libraryID,
		Deleted:          false,
		FileHash:         b.FileHash,
		FileLastModified: b.UpdatedAt,
		Oneshot:          false,
		Media: api.MediaDto{
			Comment:              "",
			EpubDivinaCompatible: false,
			EpubIsKepub:          false,
			MediaProfile:         "",
			MediaType:            "application/pdf",
			PagesCount:           int32(b.PageCount),
			Status:               "READY",
		},
		Metadata: api.BookMetadataDto{
			Authors:         []api.AuthorDto{},
			AuthorsLock:     false,
			Created:         createdAt,
			ISBN:            "",
			ISBNLock:        false,
			LastModified:    b.UpdatedAt,
			Links:           []api.WebLinkDto{},
			LinksLock:       false,
			Number:          numberStr,
			NumberLock:      false,
			NumberSort:      float64(b.Number),
			NumberSortLock:  false,
			ReleaseDate:     b.ReleaseDate,
			ReleaseDateLock: false,
			Summary:         summary,
			SummaryLock:     false,
			Tags:            []string{},
			TagsLock:        false,
			Title:           b.Name,
			TitleLock:       false,
		},
		ReadProgress: nil,
	}
}

func (h *Handlers) ListBooksV1(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	pageSize := c.DefaultQuery("size", "100")
	pageNumber := c.DefaultQuery("page", "0")

	pageSizeInt, _ := strconv.Atoi(pageSize)
	pageNumberInt, _ := strconv.Atoi(pageNumber)

	req := api.PageRequest{
		Page:    pageNumberInt,
		Size:    pageSizeInt,
		Unpaged: false,
		Sort:    nil,
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
		return
	}

	books, total, err := h.bookSer.ListBySeries(c.Request.Context(), id, req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]api.BookDto, len(books))
	for i, b := range books {
		dtos[i] = bookToDto(b, series)
	}

	writePaginatedResponse(c, req, 1, 100, total, dtos)
}

func (h *Handlers) ListBooks(c *gin.Context) {
	var search struct {
		Condition map[string]interface{} `json:"condition"`
	}

	if err := c.ShouldBindJSON(&search); err != nil {
		search.Condition = make(map[string]interface{})
	}

	var req api.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}

	seriesIDStr := extractValueAsString(search, "seriesId")
	libraryIDStr := extractValueAsString(search, "libraryId")

	var books []*models.Book
	var total int
	var err error

	if seriesIDStr != "" {
		id, err := services.StringIDToInt64(seriesIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seriesId"})
			return
		}
		books, total, err = h.bookSer.ListBySeries(c.Request.Context(), id, req.Offset(), req.Limit())

	} else if libraryIDStr != "" {
		id, err := services.StringIDToInt64(libraryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid libraryId"})
			return
		}
		books, total, err = h.bookSer.ListByLibraryID(c.Request.Context(), id, req.Offset(), req.Limit())

	} else {
		books, total, err = h.bookSer.ListAll(c.Request.Context(), req.Offset(), req.Limit())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]api.BookDto, len(books))
	seriesMap := buildSeriesMap(c.Request.Context(), h.seriesSer, books)
	for i, b := range books {
		dtos[i] = bookToDto(b, seriesMap[b.SeriesID])
	}

	writePaginatedResponse(c, req, 1, 100, total, dtos)
}

func (h *Handlers) GetBook(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), book.SeriesID)
	if err != nil {
		series = nil
	}

	c.JSON(http.StatusOK, bookToDto(book, series))
}

func (h *Handlers) ListBooksOnDeck(c *gin.Context) {
	var req api.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 0
		req.Size = 20
	}

	libraryIDsParam := c.Query("library_id")
	var libraryIDs []int64
	if libraryIDsParam != "" {
		for _, idStr := range strings.Split(libraryIDsParam, ",") {
			id, err := services.StringIDToInt64(strings.TrimSpace(idStr))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
				return
			}
			libraryIDs = append(libraryIDs, id)
		}
	}

	books, total, err := h.bookSer.ListOnDeck(c.Request.Context(), libraryIDs, req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]api.BookDto, len(books))
	seriesMap := buildSeriesMap(c.Request.Context(), h.seriesSer, books)
	for i, b := range books {
		dtos[i] = bookToDto(b, seriesMap[b.SeriesID])
	}

	writePaginatedResponse(c, req, 0, 20, total, dtos)
}

func (h *Handlers) GetBookThumbnail(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	covers, err := h.coverSer.FindForBook(c.Request.Context(), book)

	if len(covers) == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cover not found"})
		return

	}
	jpegData, err := h.thumbnailSer.GetCoverThumbnail(c.Request.Context(), covers[0])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", jpegData)
}

func (h *Handlers) ListBookThumbnails(c *gin.Context) {
	id, _ := tsidParamAsInt(c, "id")
	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	var thumbs []api.ThumbnailDto
	covers, err := h.coverSer.FindForBook(c.Request.Context(), book)

	for _, cover := range covers {
		thumbs = append(thumbs, api.ThumbnailDto{
			Type:     "book",
			ID:       strconv.Itoa(int(cover.ID)),
			Selected: false,
			URL:      "/api/v1/books/" + strconv.Itoa(int(cover.ID)) + "/thumbnails/" + strconv.Itoa(int(cover.ID)),
		})
	}
	c.JSON(http.StatusOK, thumbs)
}

func (h *Handlers) ListBooksLatest(c *gin.Context) {
	req := resolvePageRequest(c)

	libraryIDsParam := c.Query("library_id")
	var libraryIDs []int64
	if libraryIDsParam != "" {
		for _, idStr := range strings.Split(libraryIDsParam, ",") {
			id, err := services.StringIDToInt64(strings.TrimSpace(idStr))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
				return
			}
			libraryIDs = append(libraryIDs, id)
		}
	}

	books, total, err := h.bookSer.ListLatest(c.Request.Context(), libraryIDs, req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]api.BookDto, len(books))
	seriesMap := buildSeriesMap(c.Request.Context(), h.seriesSer, books)
	for i, b := range books {
		dtos[i] = bookToDto(b, seriesMap[b.SeriesID])
	}

	writePaginatedResponse(c, req, 0, 20, total, dtos)
}

func (h *Handlers) DownloadBook(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	if len(book.Episodes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book has no episodes"})
		return
	}

	var exportPages []scanApi.ExportPage
	for _, ep := range book.Episodes {
		exportPages = append(exportPages, scanApi.ExportPage{
			Filename:    ep.Filename,
			IssueNumber: ep.IssueNumber,
			Title:       ep.Title,
			PageFrom:    ep.PageFrom,
			PageTo:      ep.PageTo,
		})
	}

	tmp, err := os.CreateTemp("", "book-*.pdf")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err := scan.Build(c.Request.Context(), exportPages, false, tmp.Name()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/pdf", data)
}

func (h *Handlers) GetBookSiblingNext(c *gin.Context) {
	h.getBookSibling(c, h.bookSer.GetNextBook)
}

func (h *Handlers) GetBookSiblingPrevious(c *gin.Context) {
	h.getBookSibling(c, h.bookSer.GetPreviousBook)
}

func (h *Handlers) getBookSibling(c *gin.Context, lookup func(context.Context, int64) (*models.Book, error)) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := lookup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), book.SeriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	c.JSON(http.StatusOK, bookToDto(book, series))
}

func (h *Handlers) MarkBookReadProgress(c *gin.Context) {
	// Stubbed endpoint - accepts read progress update but doesn't persist
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Verify book exists
	_, err = h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	// Accept the request but don't persist anything
	c.Status(http.StatusNoContent)
}

func (h *Handlers) DeleteBookReadProgress(c *gin.Context) {
	// Stubbed endpoint - accepts deletion request but doesn't persist
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Verify book exists
	_, err = h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	// Accept the request but don't persist anything
	c.Status(http.StatusNoContent)
}

func buildSeriesMap(ctx context.Context, seriesSer *services.SeriesService, books []*models.Book) map[int64]*models.Series {
	m := make(map[int64]*models.Series, len(books))
	for _, b := range books {
		if _, ok := m[b.SeriesID]; ok {
			continue
		}
		s, err := seriesSer.GetByID(ctx, b.SeriesID)
		if err == nil {
			m[b.SeriesID] = s
		}
	}
	return m
}
