package vntransx

import "strings"

// VnToLatinLower maps Vietnamese accented lowercase runes to unaccented Latin equivalents.
var VnToLatinLower = map[rune]string{
	'à': "a", 'á': "a", 'ả': "a", 'ã': "a", 'ạ': "a",
	'ầ': "a", 'ấ': "a", 'ẩ': "a", 'ẫ': "a", 'ậ': "a",
	'è': "e", 'é': "e", 'ẻ': "e", 'ẽ': "e", 'ẹ': "e",
	'ề': "e", 'ế': "e", 'ể': "e", 'ễ': "e", 'ệ': "e",
	'ì': "i", 'í': "i", 'ỉ': "i", 'ĩ': "i", 'ị': "i",
	'ò': "o", 'ó': "o", 'ỏ': "o", 'õ': "o", 'ọ': "o",
	'ồ': "o", 'ố': "o", 'ổ': "o", 'ỗ': "o", 'ộ': "o",
	'ờ': "o", 'ớ': "o", 'ở': "o", 'ỡ': "o", 'ợ': "o",
	'ù': "u", 'ú': "u", 'ủ': "u", 'ũ': "u", 'ụ': "u",
	'ừ': "u", 'ứ': "u", 'ử': "u", 'ữ': "u", 'ự': "u",
	'ỳ': "y", 'ý': "y", 'ỷ': "y", 'ỹ': "y", 'ỵ': "y",
	'đ': "d", 'ê': "e", 'ă': "a", 'ơ': "o", 'ắ': "a",
	'â': "a", 'ư': "u", 'ô': "o", 'Ô': "o", 'Ồ': "o",
	'Ố': "o", 'Ổ': "o", 'Ỗ': "o", 'Ộ': "o",
}

// VnToLatinOriginal maps Vietnamese accented runes (both uppercase and lowercase) to Latin equivalents, preserving case.
var VnToLatinOriginal = map[rune]string{
	// Lowercase letters
	'à': "a", 'á': "a", 'ạ': "a", 'ả': "a", 'ã': "a", 'â': "a", 'ầ': "a", 'ấ': "a", 'ậ': "a", 'ẩ': "a", 'ẫ': "a", 'ă': "a", 'ằ': "a", 'ắ': "a", 'ặ': "a", 'ẳ': "a", 'ẵ': "a",
	'è': "e", 'é': "e", 'ẹ': "e", 'ẻ': "e", 'ẽ': "e", 'ê': "e", 'ề': "e", 'ế': "e", 'ệ': "e", 'ể': "e", 'ễ': "e",
	'ì': "i", 'í': "i", 'ị': "i", 'ỉ': "i", 'ĩ': "i",
	'ò': "o", 'ó': "o", 'ọ': "o", 'ỏ': "o", 'õ': "o", 'ô': "o", 'ồ': "o", 'ố': "o", 'ộ': "o", 'ổ': "o", 'ỗ': "o", 'ơ': "o", 'ờ': "o", 'ớ': "o", 'ợ': "o", 'ở': "o", 'ỡ': "o",
	'ù': "u", 'ú': "u", 'ụ': "u", 'ủ': "u", 'ũ': "u", 'ư': "u", 'ừ': "u", 'ứ': "u", 'ự': "u", 'ử': "u", 'ữ': "u",
	'ỳ': "y", 'ý': "y", 'ỵ': "y", 'ỷ': "y", 'ỹ': "y",
	'đ': "d",

	// Uppercase letters
	'À': "A", 'Á': "A", 'Ạ': "A", 'Ả': "A", 'Ã': "A", 'Â': "A", 'Ầ': "A", 'Ấ': "A", 'Ậ': "A", 'Ẩ': "A", 'Ẫ': "A", 'Ă': "A", 'Ằ': "A", 'Ắ': "A", 'Ặ': "A", 'Ẳ': "A", 'Ẵ': "A",
	'È': "E", 'É': "E", 'Ẹ': "E", 'Ẻ': "E", 'Ẽ': "E", 'Ê': "E", 'Ề': "E", 'Ế': "E", 'Ệ': "E", 'Ể': "E", 'Ễ': "E",
	'Ì': "I", 'Í': "I", 'Ị': "I", 'Ỉ': "I", 'Ĩ': "I",
	'Ò': "O", 'Ó': "O", 'Ọ': "O", 'Ỏ': "O", 'Õ': "O", 'Ô': "O", 'Ồ': "O", 'Ố': "O", 'Ộ': "O", 'Ổ': "O", 'Ỗ': "O", 'Ơ': "O", 'Ờ': "O", 'Ớ': "O", 'Ợ': "O", 'Ở': "O", 'Ỡ': "O",
	'Ù': "U", 'Ú': "U", 'Ụ': "U", 'Ủ': "U", 'Ũ': "U", 'Ư': "U", 'Ừ': "U", 'Ứ': "U", 'Ự': "U", 'Ử': "U", 'Ữ': "U",
	'Ỳ': "Y", 'Ý': "Y", 'Ỵ': "Y", 'Ỷ': "Y", 'Ỹ': "Y",
	'Đ': "D",
}

// ToLatin converts the input string by replacing Vietnamese accented lowercase letters
// with their unaccented Latin equivalents.
func ToLatin(input string) string {
	return convert(input, VnToLatinLower)
}

// ToLatinOriginal converts the input string by replacing Vietnamese accented letters
// (both lowercase and uppercase) with their unaccented Latin equivalents, preserving case.
func ToLatinOriginal(input string) string {
	return convert(input, VnToLatinOriginal)
}

// convert replaces runes in the input string according to the provided character map.
func convert(input string, chars map[rune]string) string {
	var result strings.Builder
	for _, r := range input {
		if replacement, ok := chars[r]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
