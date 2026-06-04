package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/chooban/progger/database"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func seriesToDto(s *database.Series) SeriesDto {
	return SeriesDto{
		ID:                   database.Int64ToStringID(s.ID),
		LibraryID:            database.Int64ToStringID(s.LibraryID),
		Name:                 s.Name,
		BookCount:            s.BookCount,
		BooksInProgressCount: 0,
		BooksReadCount:       0,
		BooksUnreadCount:     s.BookCount,
		BooksMetadata:        BookMetadataAggregationDto{},
		URL:                  s.URL(),
		Created:              s.CreatedAt,
		LastModified:         s.UpdatedAt,
		FileLastModified:     s.UpdatedAt,
		Metadata: SeriesMetadataDto{
			Title:                s.Name,
			Summary:              "",
			AgeRating:            0,
			AgeRatingLock:        false,
			AlternateTitles:      []AlternateTitleDto{},
			AlternateTitlesLock:  false,
			Created:              s.CreatedAt,
			Genres:               []string{},
			GenresLock:           false,
			Language:             "",
			LanguageLock:         false,
			LastModified:         s.UpdatedAt,
			Links:                []WebLinkDto{},
			LinksLock:            false,
			Publisher:            "",
			PublisherLock:        false,
			ReadingDirection:     "",
			ReadingDirectionLock: false,
			SharingLabels:        []string{},
			SharingLabelsLock:    false,
			Status:               "",
			StatusLock:           false,
			Tags:                 []string{},
			TagsLock:             false,
			TitleSort:            s.Name,
			TitleSortLock:        false,
			TotalBookCount:       int32(s.BookCount),
			TotalBookCountLock:   false,
		},
		Deleted: false,
		Oneshot: false,
	}
}

func (h *Handlers) RecentlyUpdatedSeries(c *gin.Context) {
	req := resolvePageRequest(c)

	series, total, err := h.seriesSer.ListRecentlyUpdated(c.Request.Context(), req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]SeriesDto, len(series))
	for i, s := range series {
		dtos[i] = seriesToDto(&s)
	}

	writePaginatedResponse(c, req, 1, 100, total, dtos)
}

func (h *Handlers) RecentlyAddedSeries(c *gin.Context) {
	req := resolvePageRequest(c)

	series, total, err := h.seriesSer.ListRecentlyAdded(c.Request.Context(), req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]SeriesDto, len(series))
	for i, s := range series {
		dtos[i] = seriesToDto(&s)
	}

	writePaginatedResponse(c, req, 1, 100, total, dtos)
}

func (h *Handlers) ListSeries(c *gin.Context) {
	var search SeriesSearchRequest

	if err := c.ShouldBindJSON(&search); err != nil && c.Request.Method == "POST" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateSeriesCondition(search.Condition); err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	req := resolvePageRequest(c)

	var series []database.Series
	var total int
	var err error

	cond := search.Condition
	switch {
	case cond.LibraryId != nil:
		id, parseErr := database.StringIDToInt64(cond.LibraryId.Value)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid libraryId"})
			return
		}
		series, total, err = h.seriesSer.ListByLibraryID(c.Request.Context(), id, req.Offset(), req.Limit())

	default:
		series, total, err = h.seriesSer.List(c.Request.Context(), req.Offset(), req.Limit())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]SeriesDto, len(series))
	for i, s := range series {
		dtos[i] = seriesToDto(&s)
	}

	writePaginatedResponse(c, req, 1, 100, total, dtos)
}

func validateSeriesCondition(cond SeriesSearchCondition) string {
	for _, v := range cond.AllOf {
		if err := validateSeriesCondition(v); err != "" {
			return err
		}
	}
	for _, v := range cond.AnyOf {
		if err := validateSeriesCondition(v); err != "" {
			return err
		}
	}
	if cond.LibraryId != nil && cond.LibraryId.Operator != "" && cond.LibraryId.Operator != "is" && cond.LibraryId.Operator != "isNot" {
		return "invalid operator for libraryId: " + cond.LibraryId.Operator
	}
	return ""
}

func (h *Handlers) GetSeries(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	c.JSON(http.StatusOK, seriesToDto(series))
}

func (h *Handlers) ListSeriesThumbnails(c *gin.Context) {
	logger, err := logr.FromContext(c)
	if err != nil {
		println("No logger found")
	}
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	series, err := h.seriesSer.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Info("No series found for id", "seriesID", id, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}
	logger.Info("Looking for all covers for series", "series", series.Name)
	covers, err := h.coverSer.FindAllBySeries(c, series.ID)
	if err != nil {
		logger.Error(err, "Failed to find covers for series", "series", series.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find covers for series"})
		return
	}
	thumbnails := make([]ThumbnailDto, len(covers))
	for i, cover := range covers {
		coverIDStr := database.Int64ToStringID(cover.ID)
		thumbnails[i] = ThumbnailDto{
			Type: "series",
			ID:   coverIDStr,
			URL:  "/api/v1/series/" + c.Param("id") + "/thumbnails/" + coverIDStr,
		}
	}
	c.JSON(http.StatusOK, thumbnails)
}

func (h *Handlers) GetSeriesThumbnailById(c *gin.Context) {
	seriesID, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid series id"})
		return
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	tIdStr := c.Param("tid")
	thumbnailID, err := strconv.ParseInt(tIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid thumbnail id"})
		return
	}

	cover, err := h.coverSer.GetByID(c.Request.Context(), thumbnailID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover found for series"})
		return
	}

	if cover.SeriesID != seriesID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover not found for series"})
		return
	}

	jpegData, err := h.thumbnailSer.GetSeriesThumbnail(c.Request.Context(), cover, series.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", jpegData)
}

func (h *Handlers) GetSeriesThumbnail(c *gin.Context) {
	seriesID, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	series, err := h.seriesSer.GetByID(c.Request.Context(), seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	cover, err := h.coverSer.FindBySeriesName(c.Request.Context(), series.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no cover found for series"})
		return
	}

	jpegData, err := h.thumbnailSer.GetSeriesThumbnail(c.Request.Context(), cover, series.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", jpegData)
}

func (h *Handlers) ListSeriesLatest(c *gin.Context) {
	req := resolvePageRequest(c)

	libraryIDsParam := c.Query("library_id")
	var libraryIDs []int64
	if libraryIDsParam != "" {
		for _, idStr := range strings.Split(libraryIDsParam, ",") {
			id, err := database.StringIDToInt64(strings.TrimSpace(idStr))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
				return
			}
			libraryIDs = append(libraryIDs, id)
		}
	}

	series, total, err := h.seriesSer.ListLatest(c.Request.Context(), libraryIDs, req.Offset(), req.Limit())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]SeriesDto, len(series))
	for i, s := range series {
		dtos[i] = seriesToDto(&s)
	}

	writePaginatedResponse(c, req, 0, 20, total, dtos)
}
