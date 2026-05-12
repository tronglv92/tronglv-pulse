package randx

import (
	"crypto/rand"
	"math/big"
)

// RandomString generates a cryptographically secure random string of the given length.
// The charset includes letters (upper and lower case), digits, underscore, and hyphen.
func RandomString(length int) (string, error) {
	var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-0123456789"
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		bytes[i] = charset[num.Int64()]
	}
	return string(bytes), nil
}
