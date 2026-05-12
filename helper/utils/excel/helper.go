package excel

import (
	"bytes"
	"pulse/helper/utils/server/http/request"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"strings"
)

// PadExcelRow ensures the row has exactly 'length' elements.
// Missing elements are filled with empty strings.
func PadExcelRow(row []string, length int) []string {
	if len(row) >= length {
		return row
	}
	padded := make([]string, length)
	copy(padded, row)
	return padded
}

// MatchExcelHeaders checks whether the given headers match the expected headers exactly,
// trimming whitespace before comparison.
func MatchExcelHeaders(headers, expected []string) bool {
	if len(headers) != len(expected) {
		return false
	}
	for i := range headers {
		if strings.TrimSpace(headers[i]) != expected[i] {
			return false
		}
	}
	return true
}

// ReadExcelFileFromRequest reads and parses an Excel file from the HTTP request under the specified form key.
// Returns an *excelize.File or an error if the file is invalid or missing.
func ReadExcelFileFromRequest(r *http.Request, key string) (*excelize.File, error) {
	uploader := request.NewFileUpload(r)

	files, err := uploader.GetFiles(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploaded file: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for key %q", key)
	}
	defer files[0].FlushData()

	fileData := files[0].GetData()

	excelFile, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	return excelFile, nil
}
