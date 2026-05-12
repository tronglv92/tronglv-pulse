package model

import (
	pmcpb "pulse/helper/utils/pmcpb/protobuf"
	"math"
)

type Pagination interface {
	GetPage() int32
	GetLimit() int32
	GetTotalRecords() int64
	GetTotalPage() int32
}

type Paginator interface {
	SetPageNumber(page int)
	SetPageSize(size int)
	PageNumber() int
	PageSize() int
	Offset() int
	WithTotalCount(total int64)
	ToPagination() Pagination
}

type paginator struct {
	page       int
	pageSize   int
	totalCount int64
}

func NewPaginator(page, size int) Paginator {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return &paginator{page: page, pageSize: size}
}

func NewSimplePaginator(total int64, size int) Paginator {
	if total <= 0 || size <= 0 {
		return &paginator{}
	}
	return &paginator{
		totalCount: total,
		pageSize:   size,
		page:       1,
	}
}

func (p *paginator) SetPageNumber(page int) {
	if page > 0 {
		p.page = page
	}
}

func (p *paginator) SetPageSize(size int) {
	if size > 0 {
		p.pageSize = size
	}
}

func (p *paginator) PageNumber() int {
	return p.page
}

func (p *paginator) PageSize() int {
	return p.pageSize
}

func (p *paginator) Offset() int {
	return (p.page - 1) * p.pageSize
}

func (p *paginator) Limit() int {
	return p.pageSize
}

func (p *paginator) WithTotalCount(total int64) {
	p.totalCount = total
}

func (p *paginator) ToPagination() Pagination {
	totalPages := 0
	if p.pageSize > 0 {
		totalPages = int(math.Ceil(float64(p.totalCount) / float64(p.pageSize)))
	}
	return &pmcpb.Pagination{
		Page:         int32(p.page),
		Limit:        int32(p.pageSize),
		TotalPage:    int32(totalPages),
		TotalRecords: p.totalCount,
	}
}

type Cursor struct {
	TotalRecords int64  `json:"total_records"`
	Limit        int    `json:"limit"`
	Next         string `json:"next"`
	Prev         string `json:"prev"`
}
