package excel

import (
	"bytes"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

const DefaultSheetName = "Sheet1"

type (
	ExcelWriter struct {
		file *excelize.File
		*Option
	}

	Header struct {
		Name  string
		Width float64
	}

	Option struct {
		HeaderHeight float64
	}

	Opt func(s *Option)
)

func NewExcelWriter(opts ...Opt) *ExcelWriter {
	w := &ExcelWriter{
		file:   excelize.NewFile(),
		Option: &Option{},
	}
	for _, opt := range opts {
		opt(w.Option)
	}
	return w
}

func (s *ExcelWriter) File() *excelize.File {
	return s.file
}

func (s *ExcelWriter) SheetName() string {
	return DefaultSheetName
}

func (s *ExcelWriter) HeaderStyle() int {
	headerStyle, err := s.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF", // White text color
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"}, // Background color (blue)
			Pattern: 1,                   // Solid fill
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		logx.Error(err)
	}
	return headerStyle
}

func (s *ExcelWriter) CoordinatesToCellName(col, row int) string {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		logx.Error(err)
	}
	return cell
}

func (s *ExcelWriter) SetHeader(sw *excelize.StreamWriter, headers []Header, styleId int, opts ...excelize.RowOpts) error {
	headerRow := make([]any, len(headers))
	for i, h := range headers {
		if err := sw.SetColWidth(i+1, i+1, h.Width); err != nil {
			return err
		}
		headerRow[i] = excelize.Cell{
			StyleID: styleId,
			Value:   h.Name,
		}
	}
	if len(opts) == 0 {
		opts = append(opts, excelize.RowOpts{Height: s.HeaderHeight})
	}
	return sw.SetRow("A1", headerRow, opts...)
}

func (s *ExcelWriter) SetHeaderRow(sw *excelize.StreamWriter, rowNumber int, rowData []any, opts ...excelize.RowOpts) error {
	opts = append(opts, excelize.RowOpts{StyleID: s.HeaderStyle()})
	return s.SetRow(sw, rowNumber, rowData, opts...)
}

func (s *ExcelWriter) SetRow(sw *excelize.StreamWriter, rowNumber int, rowData []any, opts ...excelize.RowOpts) error {
	return sw.SetRow(
		s.CoordinatesToCellName(1, rowNumber),
		rowData,
		opts...,
	)
}

func (s *ExcelWriter) WriteToBuffer() (*bytes.Buffer, error) {
	return s.file.WriteToBuffer()
}

func WithHeaderHeight(height float64) Opt {
	return func(m *Option) {
		m.HeaderHeight = height
	}
}
