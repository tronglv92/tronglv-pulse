package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"pulse/helper/utils/toolkit/jsonx"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"
)

// Md5Hash returns the MD5 hash of the input string as a hex string.
func Md5Hash(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// GetHash returns the MD5 hash of the JSON representation of a value.
func GetHash(v any) string {
	return Md5Hash(jsonx.MustMarshalString(v))
}

// Encrypt encrypts data with AES-GCM using the MD5 hash of the secret as the key.
// Returns the ciphertext with the nonce prepended.
func Encrypt(data []byte, secret string) ([]byte, error) {
	key := []byte(Md5Hash(secret))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts AES-GCM encrypted data with the MD5 hash of the secret as the key.
// Expects the nonce to be prepended to the ciphertext.
func Decrypt(data []byte, secret string) ([]byte, error) {
	key := []byte(Md5Hash(secret))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, io.ErrUnexpectedEOF
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func ParseRSAKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nb)

	e := 65537
	if len(eb) > 0 {
		e = 0
		for _, b := range eb {
			e = e<<8 + int(b)
		}
	}
	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
