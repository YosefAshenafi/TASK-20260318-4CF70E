package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const maxPageSize = 100

// ParsePagination reads page and pageSize query params (1-based page).
func ParsePagination(c *gin.Context) (page, pageSize, offset int) {
	page = 1
	pageSize = 20
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			pageSize = n
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset = (page - 1) * pageSize
	return page, pageSize, offset
}
