package server

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func (h *Handlers) ListPages(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())

	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Error(err, "failed to get book for page list")
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	pages, err := h.pageSer.GetPages(c.Request.Context(), book.ID)
	if err != nil {
		logger.Error(err, "failed to list pages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]Page, len(pages))
	for i, p := range pages {
		dtos[i] = Page{
			Number:    p.Number,
			MediaType: p.MediaType,
			FileName:  p.FileName,
			Height:    p.Height,
			Width:     p.Width,
			Size:      p.Size,
			SizeBytes: p.SizeBytes,
		}
	}

	c.JSON(http.StatusOK, dtos)
}

func (h *Handlers) GetPage(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())

	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		logger.Error(err, "bad book id for GetPage")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	pageNum, err := paramAsInt(c, "number")
	if err != nil || pageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	convert := c.Query("convert")
	accept := c.GetHeader("Accept")

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Error(err, "failed to get book for GetPage")
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	switch {
	case convert == "jpeg":
		h.servePageAsImage(c, book.ID, int(pageNum), "image/jpeg", 95)
	case convert == "png":
		h.servePageAsImage(c, book.ID, int(pageNum), "image/png", 0)
	case acceptsPDF(accept):
		data, err := h.pageSer.GetPage(c.Request.Context(), book.ID, int(pageNum))
		if err != nil {
			logger.Error(err, "failed to get PDF page")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/pdf", data)
	default:
		h.servePageAsImage(c, book.ID, int(pageNum), "image/jpeg", 95)
	}
}

func acceptsPDF(accept string) bool {
	return strings.Contains(accept, "application/pdf") || strings.Contains(accept, "*/*")
}

func (h *Handlers) servePageAsImage(c *gin.Context, bookID int64, pageNum int, contentType string, jpegQuality int) {
	logger := logr.FromContextOrDiscard(c.Request.Context())

	img, err := h.pageSer.GetPageImage(c.Request.Context(), bookID, pageNum)
	if err != nil {
		logger.Error(err, "failed to get page image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	buf := new(bytes.Buffer)
	switch contentType {
	case "image/jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: jpegQuality})
	case "image/png":
		err = png.Encode(buf, img)
	}
	if err != nil {
		logger.Error(err, "failed to encode page image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, contentType, buf.Bytes())
}

func (h *Handlers) GetPageThumbnail(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())

	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	pageNum, err := paramAsInt(c, "number")
	if err != nil || pageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	book, err := h.bookSer.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Error(err, "failed to get book for page thumbnail")
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	data, err := h.thumbnailSer.GetPageThumbnail(c.Request.Context(), book.ID, int(pageNum))
	if err != nil {
		logger.Error(err, "failed to get page thumbnail", "bookID", book.ID, "pageNum", pageNum)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}
