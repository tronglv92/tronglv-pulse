package barcode

import (
	"strings"
	"unicode"
)

type Barcode struct {
	Code        string
	BatchNumber string
	ExpiredDate string
}

func (b Barcode) GetCode() string {
	return b.Code
}

func (b Barcode) IsBarcode() bool {
	return !IsSKU(b.Code)
}

func (b Barcode) GetBatchNumber() string {
	return b.BatchNumber
}

func (b Barcode) GetExpiredDate() string {
	return b.ExpiredDate
}

func Parse(s string) Barcode {
	s = strings.TrimSpace(s)
	if s == "" {
		return Barcode{}
	}

	parts := strings.SplitN(s, ";", 3)
	b := Barcode{
		Code: parts[0],
	}

	if len(parts) > 1 {
		b.BatchNumber = parts[1]
	}
	if len(parts) > 2 {
		b.ExpiredDate = parts[2]
	}
	return b
}

func BuildBarcode(sku, batch, expireDate string) string {
	parts := []string{sku}

	if batch != "" {
		parts = append(parts, batch)
	}

	if expireDate != "" {
		parts = append(parts, expireDate)
	}

	return strings.Join(parts, ";")
}

// IsSKU checks if a given string follows SKU format:
// - exactly 6 characters
// - first character is a letter
// - next 5 characters are digits
// Example: "P29565"
func IsSKU(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}

	runes := []rune(code)
	if !unicode.IsLetter(runes[0]) {
		return false
	}

	for _, r := range runes[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
