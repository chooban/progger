package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListCollections(c *gin.Context) {
	var req PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}

	c.JSON(http.StatusOK, NewPageResponse([]CollectionDto{}, 0, req.Page, req.Size))
}

func (h *Handlers) ListReadLists(c *gin.Context) {
	var req PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.Size = 100
	}

	c.JSON(http.StatusOK, NewPageResponse([]ReadListDto{}, 0, req.Page, req.Size))
}
