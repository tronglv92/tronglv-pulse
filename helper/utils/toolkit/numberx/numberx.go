package numberx

import (
	"pulse/helper/utils/toolkit/stringx"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseInt removes non-digit characters and converts the result to int.
// Returns 0 on error or if the input is empty.
func ParseInt(s string) int {
	if stringx.IsEmptyString(s) {
		return 0
	}
	digits := removeNonDigits(s)
	i, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return i
}

// ParseInt32 removes non-digit characters and converts the result to int32.
// Returns 0 on error or if the input is empty.
func ParseInt32(s string) int32 {
	if stringx.IsEmptyString(s) {
		return 0
	}
	digits := removeNonDigits(s)
	i, err := strconv.ParseInt(digits, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i)
}

// ParseFloat converts a string to float64.
// Returns 0 on error or if the input is empty.
func ParseFloat(s string) float64 {
	if stringx.IsEmptyString(s) {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ParseFloatLocale converts a human-formatted numeric string into float64.
//
// It supports different locale formats such as:
//   - "1,234.56"  (US format)
//   - "1.234,56"  (EU format)
//   - "1234,56"   (comma as decimal separator)
//   - "1.234.567" (thousands separator)
//
// The function automatically detects the correct decimal separator and
// normalizes the input before parsing. If the input is empty or invalid,
// it returns 0 and logs the error.
func ParseFloatLocale(s string) float64 {
	s = strings.TrimSpace(s)
	if stringx.IsEmptyString(s) {
		return 0
	}

	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		if strings.LastIndex(s, ",") > strings.LastIndex(s, ".") {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	} else if strings.Contains(s, ",") {
		if strings.Count(s, ",") == 1 && len(s)-strings.LastIndex(s, ",") <= 3 {
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	} else {
		s = strings.ReplaceAll(s, ".", "")
	}

	return ParseFloat(s)
}

// SplitToInts splits a delimited string and converts each element to int.
// Returns an empty slice if input is empty or all values are invalid.
func SplitToInts(input string, delimiter string) []int {
	if stringx.IsEmptyString(input) {
		return nil
	}
	parts := strings.Split(input, delimiter)
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if i, err := strconv.Atoi(part); err == nil {
			result = append(result, i)
		}
	}
	return result
}

// removeNonDigits strips all non-digit characters from a string.
func removeNonDigits(s string) string {
	return regexp.MustCompile(`\D+`).ReplaceAllString(s, "")
}

func FormatCurrency(i int) string {
	if i < 0 {
		return "-" + FormatCurrency(-i)
	}
	if i < 1000 {
		return fmt.Sprintf("%d", i)
	}
	return FormatCurrency(i/1000) + "," + fmt.Sprintf("%03d", i%1000)
}
