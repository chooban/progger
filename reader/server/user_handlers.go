package server

import (
	"net/http"

	"github.com/chooban/progger/reader/api"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListGlobalClientSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handlers) ListUserClientSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handlers) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, api.UserDto{
		Email:              "admin@example.com",
		ID:                 "1",
		LabelsAllow:        []string{},
		LabelsExclude:      []string{},
		Roles:              []string{"ADMIN"},
		SharedAllLibraries: false,
		SharedLibrariesIds: []string{},
		AgeRestriction:     nil,
	})
}
