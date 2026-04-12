package dt

import (
	"math"
	"time"
)

const (
	OrderASC  OrderDirection = "asc"
	OrderDESC OrderDirection = "desc"
)

type (
	UserID         string
	OrderDirection string
)

type BaseModel struct {
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type PagingParam struct {
	WithCount bool `form:"with_count" json:"with_count"`
	Page      uint `form:"page" json:"page" `
	Limit     uint `form:"limit" json:"limit"`
}

func (p PagingParam) ToResponse(totalItems uint64) PagingResponse {
	totalPage := 0

	if p.Limit > 0 {
		totalPage = int(math.Ceil(float64(totalItems) / float64(p.Limit)))
	}

	if p.Page <= 0 {
		p.Page = 1
	}

	return PagingResponse{
		CurrentPage: p.Page,
		PageSize:    p.Limit,
		TotalPage:   uint(totalPage),
		TotalItems:  totalItems,
	}
}

type PagingResponse struct {
	CurrentPage uint   `json:"current_page"`
	PageSize    uint   `json:"page_size"`
	TotalPage   uint   `json:"total_page"`
	TotalItems  uint64 `json:"total_items"`
}

type Paginated[T any] struct {
	PagingResponse
	Items []T `json:"items"`
}
