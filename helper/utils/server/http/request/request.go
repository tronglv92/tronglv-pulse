package request

type SortOrderReq struct {
	SortBy    string `form:"sort_by,default=id"`
	SortOrder string `form:"sort_order,default=DESC"`
}

func (req *SortOrderReq) GetSortBy() string {
	return req.SortBy
}

func (req *SortOrderReq) GetSortOrder() string {
	return req.SortOrder
}

type PaginationReq struct {
	Limit int    `form:"limit,default=10"`
	Page  int    `form:"page,default=1"`
	Prev  string `form:"prev,optional"`
	Next  string `form:"next,optional"`
}

func (req *PaginationReq) GetLimit() int {
	return req.Limit
}

func (req *PaginationReq) GetPage() int {
	return req.Page
}

func (req *PaginationReq) GetPrev() string {
	return req.Prev
}

func (req *PaginationReq) GetNext() string {
	return req.Next
}
