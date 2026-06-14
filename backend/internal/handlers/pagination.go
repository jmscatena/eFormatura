package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaginateParams contains pagination parameters
type PaginateParams struct {
	Page    int
	Limit   int
	OrderBy string
}

// PaginateResponse wraps paginated results with metadata
type PaginateResponse struct {
	Data     interface{} `json:"data"`
	Page     int         `json:"page"`
	Limit    int         `json:"limit"`
	Total    int64       `json:"total"`
	TotalPages int      `json:"total_pages"`
}

// ParsePaginateParams extracts pagination parameters from context
func ParsePaginateParams(c *gin.Context) PaginateParams {
	page, _ := c.GetQuery("page")
	limit, _ := c.GetQuery("limit")
	orderBy := c.DefaultQuery("order_by", "created_at")

	p := PaginateParams{
		Page:    1,
		Limit:   50,
		OrderBy: orderBy,
	}

	// Parse page
	if p2, err := pageToInt(page); err == nil && p2 > 0 {
		p.Page = p2
	}

	// Parse limit
	if l, err := pageToInt(limit); err == nil && l > 0 && l <= 200 {
		p.Limit = l
	}

	return p
}

// pageToInt converts string to int, returns 0 on error
func pageToInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// Paginate applies pagination to a GORM query
func Paginate(db *gorm.DB, params PaginateParams, model interface{}) PaginateResponse {
	offset := (params.Page - 1) * params.Limit

	// Apply ordering
	if params.OrderBy != "" {
		db = db.Order(params.OrderBy)
	}

	// Count total records
	var total int64
	db.Model(model).Count(&total)

	// Apply pagination and fetch records
	db = db.Offset(offset).Limit(params.Limit)
	db.Find(model)

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return PaginateResponse{
		Data:       model,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
