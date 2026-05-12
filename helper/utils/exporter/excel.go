package exporter

import (
	"bytes"
	"context"
	"pulse/helper/utils/excel"
	"fmt"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	ExcelHandler interface {
		Build(ctx context.Context, writer *excel.ExcelWriter) (*excelize.File, error)
	}

	excelExporter struct {
		fileName string
		handler  ExcelHandler
	}
)

func NewExcelExporter(fileName string, handler ExcelHandler) Exporter {
	return &excelExporter{
		fileName: fileName,
		handler:  handler,
	}
}

func (e *excelExporter) FileName() string {
	return e.fileName
}

func (e *excelExporter) Export(ctx context.Context, w Writer) (*bytes.Buffer, error) {
	excelWriter, ok := w.(*excel.ExcelWriter)
	if !ok {
		return nil, fmt.Errorf("w is not a ExcelWriter")
	}
	f, err := e.handler.Build(ctx, excelWriter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = f.Close(); err != nil {
			logx.Error(err)
		}
	}()
	return f.WriteToBuffer()
}
