package server

import (
	"net/http"
	"strconv"

	"github.com/chooban/progger/database"
	"github.com/gin-gonic/gin"
)

type viewerData struct {
	Title           string
	ContentTemplate string
	Breadcrumbs     []breadcrumb
	Libraries       []readerLibrary
	Series          []readerSeries
	Books           []readerBook
	LibraryID       string
	BookID          string
	BookName        string
	CurrentPage     int
	TotalPages      int
	PrevPage        int
	NextPage        int
	PrevBookID      string
	NextBookID      string
	PageThumbnails  []pageThumbnail
}

type pageThumbnail struct {
	URL     string
	PageNum int
}

type breadcrumb struct {
	Label string
	URL   string
}

type readerLibrary struct {
	ID   string
	Name string
}

type readerSeries struct {
	ID        string
	Name      string
	BookCount int
}

type readerBook struct {
	ID        string
	Name      string
	Number    int
	PageCount int
}

func (h *Handlers) ReadLibrariesPage(c *gin.Context) {
	ctx := c.Request.Context()

	libraries, err := h.librarySer.List(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to list libraries")
		return
	}

	data := viewerData{
		Title:           "Libraries",
		ContentTemplate: "library_content",
		Libraries:       make([]readerLibrary, 0, len(libraries)),
	}
	for _, lib := range libraries {
		data.Libraries = append(data.Libraries, readerLibrary{
			ID:   database.Int64ToStringID(lib.ID),
			Name: lib.Name,
		})
	}

	c.HTML(http.StatusOK, "library_list", data)
}

func (h *Handlers) ReadSeriesPage(c *gin.Context) {
	ctx := c.Request.Context()

	libraryID, err := database.StringIDToInt64(c.Param("lid"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid library ID")
		return
	}

	library, err := h.librarySer.GetByID(ctx, libraryID)
	if err != nil {
		c.String(http.StatusNotFound, "Library not found")
		return
	}

	allSeries, _, err := h.seriesSer.List(ctx, 0, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to list series")
		return
	}

	var matching []readerSeries
	for _, s := range allSeries {
		if s.LibraryID == libraryID {
			matching = append(matching, readerSeries{
				ID:        database.Int64ToStringID(s.ID),
				Name:      s.Name,
				BookCount: s.BookCount,
			})
		}
	}

	data := viewerData{
		Title:           library.Name,
		ContentTemplate: "series_content",
		LibraryID:       database.Int64ToStringID(libraryID),
		Series:    matching,
		Breadcrumbs: []breadcrumb{
			{Label: library.Name, URL: "/read/libraries/" + database.Int64ToStringID(libraryID) + "/series"},
		},
	}

	c.HTML(http.StatusOK, "series_list", data)
}

func (h *Handlers) ReadBooksPage(c *gin.Context) {
	ctx := c.Request.Context()

	libraryID, err := database.StringIDToInt64(c.Param("lid"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid library ID")
		return
	}

	seriesID, err := database.StringIDToInt64(c.Param("sid"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid series ID")
		return
	}

	library, err := h.librarySer.GetByID(ctx, libraryID)
	if err != nil {
		c.String(http.StatusNotFound, "Library not found")
		return
	}

	series, err := h.seriesSer.GetByID(ctx, seriesID)
	if err != nil {
		c.String(http.StatusNotFound, "Series not found")
		return
	}

	books, _, err := h.bookSer.ListBySeries(ctx, seriesID, 0, 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to list books")
		return
	}

	libStr := database.Int64ToStringID(libraryID)
	seriesStr := database.Int64ToStringID(seriesID)

	var bookViews []readerBook
	for _, b := range books {
		bookViews = append(bookViews, readerBook{
			ID:        database.Int64ToStringID(b.ID),
			Name:      b.Name,
			Number:    b.Number,
			PageCount: b.PageCount,
		})
	}

	data := viewerData{
		Title:           series.Name,
		ContentTemplate: "book_content",
		LibraryID:       libStr,
		Books:     bookViews,
		Breadcrumbs: []breadcrumb{
			{Label: library.Name, URL: "/read/libraries/" + libStr + "/series"},
			{Label: series.Name, URL: "/read/libraries/" + libStr + "/series/" + seriesStr + "/books"},
		},
	}

	c.HTML(http.StatusOK, "book_list", data)
}

func (h *Handlers) ReadBookPage(c *gin.Context) {
	ctx := c.Request.Context()

	libraryID, err := database.StringIDToInt64(c.Param("lid"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid library ID")
		return
	}

	bookID, err := database.StringIDToInt64(c.Param("bid"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid book ID")
		return
	}

	pageNum := 1
	if p := c.Param("page"); p != "" {
		pageNum, err = strconv.Atoi(p)
		if err != nil || pageNum < 1 {
			pageNum = 1
		}
	}

	library, err := h.librarySer.GetByID(ctx, libraryID)
	if err != nil {
		c.String(http.StatusNotFound, "Library not found")
		return
	}

	book, err := h.bookSer.GetByID(ctx, bookID)
	if err != nil {
		c.String(http.StatusNotFound, "Book not found")
		return
	}

	series, err := h.seriesSer.GetByID(ctx, book.SeriesID)
	if err != nil {
		c.String(http.StatusNotFound, "Series not found")
		return
	}

	totalPages := computeTotalPages(book.Episodes)
	if pageNum > totalPages {
		pageNum = totalPages
	}
	if totalPages == 0 {
		pageNum = 0
	}

	libStr := database.Int64ToStringID(libraryID)
	bookStr := database.Int64ToStringID(bookID)
	seriesStr := database.Int64ToStringID(book.SeriesID)

	data := viewerData{
		Title:           book.Name + " — Page " + strconv.Itoa(pageNum),
		ContentTemplate: "viewer_content",
		BookID:          bookStr,
		LibraryID:   libStr,
		BookName:    book.Name,
		CurrentPage: pageNum,
		TotalPages:  totalPages,
		Breadcrumbs: []breadcrumb{
			{Label: library.Name, URL: "/read/libraries/" + libStr + "/series"},
			{Label: series.Name, URL: "/read/libraries/" + libStr + "/series/" + seriesStr + "/books"},
			{Label: book.Name, URL: "/read/libraries/" + libStr + "/books/" + bookStr + "/pages/1"},
		},
	}

	if pageNum > 1 {
		data.PrevPage = pageNum - 1
	}
	if pageNum < totalPages {
		data.NextPage = pageNum + 1
	}

	if pageNum == 1 {
		prevBook, err := h.bookSer.GetPreviousBook(ctx, bookID)
		if err == nil && prevBook != nil {
			data.PrevBookID = database.Int64ToStringID(prevBook.ID)
		}
	}
	if pageNum == totalPages {
		nextBook, err := h.bookSer.GetNextBook(ctx, bookID)
		if err == nil && nextBook != nil {
			data.NextBookID = database.Int64ToStringID(nextBook.ID)
		}
	}

	data.PageThumbnails = make([]pageThumbnail, totalPages)
	for i := 0; i < totalPages; i++ {
		page := i + 1
		data.PageThumbnails[i] = pageThumbnail{
			URL:     "/api/v1/books/" + bookStr + "/pages/" + strconv.Itoa(page) + "/thumbnail",
			PageNum: page,
		}
	}

	c.HTML(http.StatusOK, "viewer", data)
}

func computeTotalPages(episodes []*database.Episode) int {
	total := 0
	for _, ep := range episodes {
		total += ep.PageTo - ep.PageFrom + 1
	}
	return total
}
