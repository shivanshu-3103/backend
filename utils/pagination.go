package utils

import (
	"net/http"
	"strconv"
)

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

func GetPaginationParams(r *http.Request) (page int, limit int, offset int) {
	page = 1
	limit = 10

	query := r.URL.Query()
	
	if p, err := strconv.Atoi(query.Get("page")); err == nil && p > 0 {
		page = p
	}
	
	if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
		limit = l
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

func CalculateTotalPages(totalCount int64, limit int) int {
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}
	return totalPages
}
