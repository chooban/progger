package server

import (
	"net/http"
	"strconv"

	"github.com/chooban/progger/database"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	librarySer   *database.LibraryRepo
	seriesSer    *database.SeriesRepo
	bookSer      *database.BookRepo
	pageSer      *PageService
	coverSer     *database.CoverRepo
	thumbnailSer *ThumbnailService
	scanSer      *ScanService
}

func NewHandlers(librarySer *database.LibraryRepo, seriesSer *database.SeriesRepo, bookSer *database.BookRepo, pageSer *PageService, coverSer *database.CoverRepo, thumbnailSer *ThumbnailService, scanSer *ScanService) *Handlers {
	return &Handlers{
		librarySer:   librarySer,
		seriesSer:    seriesSer,
		bookSer:      bookSer,
		pageSer:      pageSer,
		coverSer:     coverSer,
		thumbnailSer: thumbnailSer,
		scanSer:      scanSer,
	}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "0.0.1",
	})
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func tsidParamAsInt(c *gin.Context, param string) (int64, error) {
	paramString := c.Param(param)
	if paramString == "" {
		return 0, nil
	}
	// Parse TSID string (13-char base32)
	return database.StringIDToInt64(paramString)
}

func paramAsInt(c *gin.Context, param string) (int64, error) {
	paramString := c.Param(param)
	if paramString == "" {
		return 0, nil
	}
	return strconv.ParseInt(paramString, 10, 64)
}

func writePaginatedResponse[T any](c *gin.Context, req PageRequest, defaultPage, defaultSize int, total int, dtos []T) {
	page := defaultPage
	if req.Page > 0 {
		page = req.Page
	}
	size := defaultSize
	if req.Size > 0 {
		size = req.Size
	}
	resp := NewPageResponse(dtos, total, page, size)
	c.JSON(http.StatusOK, resp)
}

func libraryToDto(lib *database.Library) LibraryDto {
	return LibraryDto{
		ID:   database.Int64ToStringID(lib.ID),
		Name: lib.Name,
	}
}

func resolvePageRequest(c *gin.Context) PageRequest {
	var req PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}
	return req
}
