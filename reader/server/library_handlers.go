package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func (h *Handlers) ListLibraries(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())
	libs, err := h.librarySer.List(c.Request.Context())
	if err != nil {
		logger.Error(err, "failed to list libraries", "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("libraries listed", "count", len(libs))
	dtos := make([]LibraryDto, len(libs))
	for i, lib := range libs {
		dtos[i] = libraryToDto(&lib)
	}

	c.JSON(http.StatusOK, dtos)
}

func (h *Handlers) GetLibrary(c *gin.Context) {
	id, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	lib, err := h.librarySer.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "library not found"})
		return
	}

	c.JSON(http.StatusOK, libraryToDto(lib))
}

func (h *Handlers) ScanLibrary(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())
	cfg := h.scanSer.Cfg()

	logger.Info("scan library requested", "scan_dirs", cfg.ScanDirectories, "library", cfg.LibraryName)

	if cfg.ScanDirectories == nil || len(cfg.ScanDirectories) == 0 {
		logger.Error(nil, "no scan directories configured")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no scan directories configured"})
		return
	}

	go func() {
		logger.Info("starting background scan")
		// TODO: Should I create a new context here? The request one is cancelled when the request returns
		ctx := logr.NewContext(context.Background(), logger)
		if err := h.scanSer.ScanDirectories(ctx); err != nil {
			logger.Error(err, "failed to scan directories")
		} else {
			logger.Info("scan directories finished")
		}
	}()

	c.Status(http.StatusAccepted)
}

func (h *Handlers) RefreshLibraryMetadata(c *gin.Context) {
	logger := logr.FromContextOrDiscard(c.Request.Context())
	libraryID, err := tsidParamAsInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library id"})
		return
	}

	go func() {
		ctx := logr.NewContext(context.Background(), logger)
		if err := h.scanSer.RefreshMetadata(ctx, libraryID); err != nil {
			logger.Error(err, "failed to refresh metadata")
		} else {
			logger.Info("metadata refresh finished")
		}
	}()

	c.Status(http.StatusAccepted)
}
