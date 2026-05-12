package model

type ExportModel struct {
	ExportToEmail      string     `json:"-"`
	ExportToEmployeeId string     `json:"-"`
	ExportType         ExportType `form:"export_type"`
}

func (s ExportModel) GetEmail() string {
	return s.ExportToEmail
}

func (s ExportModel) GetEmployeeId() string {
	return s.ExportToEmployeeId
}

func (s ExportModel) GetExportType() ExportType {
	return s.ExportType
}
