package server

import (
	"net/http"

	"github.com/chooban/progger/reader/api"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListCollections(c *gin.Context) {
	var req api.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}

	c.JSON(http.StatusOK, api.NewPageResponse([]api.CollectionDto{}, 0, req.Page, req.Size))
}

func (h *Handlers) ListReadLists(c *gin.Context) {
	var req api.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}

	c.JSON(http.StatusOK, api.NewPageResponse([]api.ReadListDto{}, 0, req.Page, req.Size))
}
