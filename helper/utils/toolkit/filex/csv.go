package filex

import (
	"encoding/csv"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"os"
)

func CsvReader(filePath string) ([]string, [][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			logx.Errorf("failed to close file: %w", err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	csvReader.LazyQuotes = true // Allow unescaped quotes
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	var header []string
	var rows [][]string

	for lineNum := 1; ; lineNum++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("CSV parsing error on line %d: %w", lineNum, err)
		}

		if len(header) == 0 {
			header = record
			continue
		}

		if len(record) != len(header) {
			logx.Infof("line %d: header has %d fields but record has %d fields", lineNum, len(header), len(record))
		}
		rows = append(rows, record)
	}

	if len(header) == 0 {
		return nil, nil, fmt.Errorf("empty CSV file")
	}

	return header, rows, nil
}
