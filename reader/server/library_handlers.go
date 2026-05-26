package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chooban/progger/reader/api"
	"github.com/chooban/progger/reader/config"
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

	logger.Info(fmt.Sprintf("Found %d libraries", len(libs)))
	dtos := make([]api.LibraryDto, len(libs))
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
	cfg := config.Load()

	if cfg.ScanDirectories == nil || len(cfg.ScanDirectories) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no scan directories configured"})
		return
	}

	go func() {
		// TODO: Should I create a new context here? The request one is cancelled when the request returns
		logger := logr.FromContextOrDiscard(c.Request.Context())
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
