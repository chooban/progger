package server

import "github.com/gin-gonic/gin"

func registerViewerRoutes(router *gin.Engine, handlers *Handlers) {
	read := router.Group("/read")
	read.GET("/", handlers.ReadLibrariesPage)
	read.GET("/libraries/:lid/series", handlers.ReadSeriesPage)
	read.GET("/libraries/:lid/series/:sid/books", handlers.ReadBooksPage)
	read.GET("/libraries/:lid/books/:bid/pages/:page", handlers.ReadBookPage)
	read.GET("/libraries/:lid/books/:bid", func(c *gin.Context) {
		c.Redirect(302, "/read/libraries/"+c.Param("lid")+"/books/"+c.Param("bid")+"/pages/1")
	})
}
