package exporter

import (
	"bytes"
	"context"
)

type Exporter interface {
	FileName() string
	Export(ctx context.Context, w Writer) (*bytes.Buffer, error)
}

type Writer interface {
	WriteToBuffer() (*bytes.Buffer, error)
}

type ExportEvent struct {
	Model      string `json:"model"`
	Email      string `json:"email"`
	EmployeeId string `json:"employee_id"`
	Data       any    `json:"data"`
}
