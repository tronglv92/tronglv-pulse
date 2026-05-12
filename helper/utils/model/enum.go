package model

type SortOrderType string

const (
	OrderAsc  SortOrderType = "asc"
	OrderDesc SortOrderType = "desc"
)

type Status int

const (
	ActiveStatus   Status = 1
	InActiveStatus Status = 2
)

type ExportType int

const (
	ExportTypeUrl    ExportType = 1
	ExportTypeDirect ExportType = 2
)
