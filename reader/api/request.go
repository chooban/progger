package api

import "strconv"

type PageRequest struct {
	Page    int      `form:"page"`
	Size    int      `form:"size"`
	Unpaged bool     `form:"unpaged"`
	Sort    []string `form:"sort"`
}

type SearchRequest struct {
	PageRequest
	SeriesID string `form:"seriesId"`
	Search   string `form:"search"`
}

func (r *PageRequest) Offset() int {
	if r.Page <= 0 {
		return 0
	}
	return (r.Page - 1) * r.Size
}

func (r *PageRequest) Limit() int {
	if r.Unpaged || r.Size <= 0 {
		return -1
	}
	return r.Size
}

type Pageable struct {
	PageNumber int64      `json:"pageNumber"`
	PageSize   int64      `json:"pageSize"`
	Offset     int64      `json:"offset"`
	Paged      bool       `json:"paged"`
	Unpaged    bool       `json:"unpaged"`
	Sort       SortObject `json:"sort"`
}

type SortObject struct {
	Empty    bool `json:"empty"`
	Sorted   bool `json:"sorted"`
	Unsorted bool `json:"unsorted"`
}

type PageResponse[T any] struct {
	Content          T          `json:"content"`
	Empty            bool       `json:"empty"`
	First            bool       `json:"first"`
	Last             bool       `json:"last"`
	Number           int32      `json:"number"`
	NumberOfElements int32      `json:"numberOfElements"`
	Pageable         Pageable   `json:"pageable"`
	Size             int32      `json:"size"`
	Sort             SortObject `json:"sort"`
	TotalElements    int64      `json:"totalElements"`
	TotalPages       int32      `json:"totalPages"`
}

func NewPageResponse[T any](content T, totalElements int, page, size int) *PageResponse[T] {
	contentSlice := any(content)
	var count int64 = 0

	if arr, ok := contentSlice.([]interface{}); ok {
		count = int64(len(arr))
	} else if sl, ok := contentSlice.([]SeriesDto); ok {
		count = int64(len(sl))
	} else if sl, ok := contentSlice.([]BookDto); ok {
		count = int64(len(sl))
	} else if sl, ok := contentSlice.([]CollectionDto); ok {
		count = int64(len(sl))
	} else if sl, ok := contentSlice.([]ReadListDto); ok {
		count = int64(len(sl))
	}

	totalPageVal := int32(0)
	if size > 0 {
		totalPageVal = int32((totalElements + size - 1) / size)
		if totalPageVal < 1 {
			totalPageVal = 1
		}
	}

	return &PageResponse[T]{
		Content:          content,
		Empty:            count == 0,
		First:            page <= 1,
		Last:             count == 0 || totalPageVal <= int32(page),
		Number:           int32(page),
		NumberOfElements: int32(count),
		Pageable:         Pageable{PageNumber: int64(page), PageSize: int64(size), Paged: true, Unpaged: false, Sort: SortObject{Empty: true, Sorted: false, Unsorted: true}},
		Size:             int32(size),
		Sort:             SortObject{Empty: true, Sorted: false, Unsorted: true},
		TotalElements:    int64(totalElements),
		TotalPages:       totalPageVal,
	}
}

func FormatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}
