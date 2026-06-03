package server

import (
	"net/http"
	"strconv"

	"github.com/chooban/progger/reader/api"
	"github.com/chooban/progger/reader/models"
	"github.com/chooban/progger/reader/services"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	librarySer   *services.LibraryService
	seriesSer    *services.SeriesService
	bookSer      *services.BookService
	pageSer      *services.PageService
	coverSer     *services.CoverService
	thumbnailSer *services.ThumbnailService
	scanSer      *services.ScanService
}

func NewHandlers(librarySer *services.LibraryService, seriesSer *services.SeriesService, bookSer *services.BookService, pageSer *services.PageService, coverSer *services.CoverService, thumbnailSer *services.ThumbnailService, scanSer *services.ScanService) *Handlers {
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
	return services.StringIDToInt64(paramString)
}

func paramAsInt(c *gin.Context, param string) (int64, error) {
	paramString := c.Param(param)
	if paramString == "" {
		return 0, nil
	}
	return strconv.ParseInt(paramString, 10, 64)
}

func writePaginatedResponse[T any](c *gin.Context, req api.PageRequest, defaultPage, defaultSize int, total int, dtos []T) {
	page := defaultPage
	if req.Page > 0 {
		page = req.Page
	}
	size := defaultSize
	if req.Size > 0 {
		size = req.Size
	}
	resp := api.NewPageResponse(dtos, total, page, size)
	c.JSON(http.StatusOK, resp)
}

func libraryToDto(lib *models.Library) api.LibraryDto {
	return api.LibraryDto{
		AnalyzeDimensions:                 lib.AnalyzeDimensions,
		ConvertToCbz:                      lib.ConvertToCbz,
		EmptyTrashAfterScan:               lib.EmptyTrashAfterScan,
		HashFiles:                         lib.HashFiles,
		HashKoreader:                      lib.HashKoreader,
		HashPages:                         lib.HashPages,
		ID:                                services.Int64ToStringID(lib.ID),
		ImportBarcodeIsbn:                 lib.ImportBarcodeIsbn,
		ImportComicInfoBook:               lib.ImportComicInfoBook,
		ImportComicInfoCollection:         lib.ImportComicInfoCollection,
		ImportComicInfoReadList:           lib.ImportComicInfoReadList,
		ImportComicInfoSeries:             lib.ImportComicInfoSeries,
		ImportComicInfoSeriesAppendVolume: lib.ImportComicInfoSeriesAppendVolume,
		ImportEpubBook:                    lib.ImportEpubBook,
		ImportEpubSeries:                  lib.ImportEpubSeries,
		ImportLocalArtwork:                lib.ImportLocalArtwork,
		ImportMylarSeries:                 lib.ImportMylarSeries,
		Name:                              lib.Name,
		OneshotsDirectory:                 lib.OneshotsDirectory,
		RepairExtensions:                  lib.RepairExtensions,
		Root:                              lib.Root,
		ScanCbx:                           lib.ScanCbx,
		ScanDirectoryExclusions:           lib.ScanDirectoryExclusions,
		ScanEpub:                          lib.ScanEpub,
		ScanForceModifiedTime:             lib.ScanForceModifiedTime,
		ScanInterval:                      lib.ScanInterval,
		ScanOnStartup:                     lib.ScanOnStartup,
		ScanPdf:                           lib.ScanPdf,
		SeriesCover:                       lib.SeriesCover,
		Unavailable:                       lib.Unavailable,
	}
}

func resolvePageRequest(c *gin.Context) api.PageRequest {
	var req api.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}
	return req
}
