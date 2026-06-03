package server

import (
	"github.com/gin-gonic/gin"
)

func ConfigureRoutes(router *gin.Engine, handlers *Handlers) {
	registerViewerRoutes(router, handlers)

	router.GET("/", handlers.HealthCheck)
	router.POST("/scan", handlers.ScanLibrary)

	public := router.Group("/api/v1")

	publicV2 := router.Group("/api/v2")
	publicV2.GET("/users/me", handlers.GetMe)

	public.GET("/libraries", handlers.ListLibraries)
	public.GET("/libraries/:id", handlers.GetLibrary)
	public.POST("/libraries/:id/scan", handlers.ScanLibrary)
	public.POST("/libraries/:id/metadata/refresh", handlers.RefreshLibraryMetadata)

	// Client settings endpoints (no auth required)
	publicNoAuth := router.Group("/api/v1")
	publicNoAuth.GET("/client-settings/global/list", handlers.ListGlobalClientSettings)
	publicNoAuth.GET("/client-settings/user/list", handlers.ListUserClientSettings)

	public.GET("/series/new", handlers.RecentlyAddedSeries)
	public.GET("/series/updated", handlers.RecentlyUpdatedSeries)
	public.GET("/series/latest", handlers.ListSeriesLatest)
	public.GET("/series", handlers.ListSeries)
	public.POST("/series/list", handlers.ListSeries)
	public.GET("/series/:id", handlers.GetSeries)
	public.GET("/series/:id/books", handlers.ListBooksV1)
	public.GET("/series/:id/thumbnails", handlers.ListSeriesThumbnails)
	public.GET("/series/:id/thumbnails/:tid", handlers.GetSeriesThumbnailById)
	public.GET("/series/:id/thumbnail", handlers.GetSeriesThumbnail)

	public.GET("/books", handlers.ListBooks)
	public.POST("/books/list", handlers.ListBooks)
	public.GET("/books/latest", handlers.ListBooksLatest)
	public.GET("/books/ondeck", handlers.ListBooksOnDeck)
	public.GET("/books/:id", handlers.GetBook)
	public.GET("/books/:id/file", handlers.DownloadBook)
	public.GET("/books/:id/file/*filename", handlers.DownloadBook)
	public.GET("/books/:id/next", handlers.GetBookSiblingNext)
	public.GET("/books/:id/previous", handlers.GetBookSiblingPrevious)
	public.PATCH("/books/:id/read-progress", handlers.MarkBookReadProgress)
	public.DELETE("/books/:id/read-progress", handlers.DeleteBookReadProgress)
	public.GET("/books/:id/pages", handlers.ListPages)
	public.GET("/books/:id/pages/:number", handlers.GetPage)
	public.GET("/books/:id/pages/:number/thumbnail", handlers.GetPageThumbnail)
	public.GET("/books/:id/thumbnail", handlers.GetBookThumbnail)
	public.GET("/books/:id/thumbnails", handlers.ListBookThumbnails)
	public.GET("/books/:id/thumbnails/:tid", handlers.GetBookThumbnail)

	public.GET("/collections", handlers.ListCollections)
	public.GET("/readlists", handlers.ListReadLists)

}
