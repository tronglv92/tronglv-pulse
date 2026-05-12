package stringx

import (
	"pulse/helper/utils/model"
	"pulse/helper/utils/toolkit/vntransx"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// EscapeHTML escapes all string fields in a struct (by pointer) using html.EscapeString.
// It modifies the input struct in-place.
func EscapeHTML(input interface{}) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return // Ignore non-pointer or non-struct inputs
	}

	val := v.Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			escaped := html.EscapeString(field.String())
			field.SetString(escaped)
		}
	}
}

// StripHTMLTags removes HTML tags and returns plain text from an HTML string.
func StripHTMLTags(input string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var result strings.Builder

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return strings.TrimSpace(result.String())
		case html.TextToken:
			text := strings.TrimSpace(tokenizer.Token().Data)
			if text != "" {
				result.WriteString(text)
				result.WriteRune(' ')
			}
		}
	}
}

// TruncateStrict limits the string to a maximum number of characters.
// Adds "..." if the string was truncated.
func TruncateStrict(text string, limit int) string {
	if text == "" || len(text) <= limit {
		return text
	}
	reader := strings.NewReader(text)
	buf := make([]byte, limit)
	n, _ := io.ReadAtLeast(reader, buf, limit)
	if n > 0 {
		return fmt.Sprintf("%s...", string(buf))
	}
	return text
}

// TruncateByWord limits the string to a maximum number of characters.
// Adds "..." if the string was truncated.
func TruncateByWord(text string, limit int) string {
	if text == "" || len(text) <= limit {
		return text
	}

	// Take a slice of bytes up to the limit
	buf := []byte(text[:limit])

	// Find the last whitespace in that slice
	lastSpace := -1
	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] == ' ' || buf[i] == '\t' || buf[i] == '\n' {
			lastSpace = i
			break
		}
	}

	if lastSpace > 0 {
		return string(buf[:lastSpace]) + "..."
	}
	return string(buf) + "..."
}

// Slugify converts a string to a URL-friendly slug using only lowercase letters, numbers, and dashes.
func Slugify(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	normalized := strings.ToLower(strings.TrimSpace(input))
	latin := toLatin(normalized)
	spaceSeparated := re.ReplaceAllString(latin, " ")
	trimmed := strings.TrimSpace(spaceSeparated)
	return strings.ReplaceAll(trimmed, " ", "-")
}

// ToCode converts a string to an identifier-friendly format (snake case with underscores).
func ToCode(s string) string {
	return strings.ReplaceAll(Slugify(s), "-", "_")
}

// ToCodeUpper converts a string to an identifier-friendly format (snake case with underscores).
func ToCodeUpper(s string) string {
	return strings.ToUpper(ToCode(s))
}

// ToSnakeCase converts CamelCase or PascalCase strings to snake_case.
func ToSnakeCase(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	return strings.ToLower(snake)
}

// Dummy placeholder for unicode.ToLatin (not in standard library).
// Replace with a real transliteration if needed.
func toLatin(s string) string {
	return vntransx.ToLatin(s)
}

// SortOrder normalizes a sort order string to "asc" or "desc".
// Returns "asc" if input is empty or invalid (case-insensitive).
func SortOrder(s string) string {
	s = strings.ToLower(s)
	if s == "" || (s != string(model.OrderAsc) && s != string(model.OrderDesc)) {
		return string(model.OrderAsc)
	}
	return s
}

// JoinIfNotEmpty joins the elements of the slice with the given separator,
// but returns an empty string if the slice is nil or empty.
func JoinIfNotEmpty(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, sep)
}

// IsNilOrEmptyString checks whether the given value is either nil or an empty string.
// It safely handles an `any` (interface{}) type by asserting whether the value is a string.
// Returns true if:
//   - The value is nil,
//   - The value is not a string,
//   - The string is empty ("").
//
// This is useful when working with dynamic types (e.g., proto custom options or generic metadata)
// where a string value might be optional or unset.
func IsNilOrEmptyString(val any) bool {
	if val == nil {
		return true
	}
	str, ok := val.(string)
	return !ok || str == ""
}

// IsEmptyString returns true if the string is empty or contains only whitespace.
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// JoinURL safely joins multiple URL path segments into a single well-formed URL.
//
// Example:
//
//	JoinURL("https://example.com/", "/api/", "/v1/", "users")
//	→ "https://example.com/api/v1/users"
//
// It automatically:
//   - Removes duplicate slashes between segments
//   - Keeps protocol slashes intact (e.g., "https://")
//   - Works with any number of path segments
func JoinURL(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	result := paths[0]
	for i := 1; i < len(paths); i++ {
		result = fmt.Sprintf(
			"%s/%s",
			strings.TrimRight(result, "/"),
			strings.TrimLeft(paths[i], "/"),
		)
	}
	return result
}
