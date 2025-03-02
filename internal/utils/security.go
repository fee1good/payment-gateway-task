package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

var (
	// todo move to config / get from vault
	secretKey = []byte("super-puper-secret-key")
)

func MaskData(data []byte) string {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		// В продакшене нужно правильно обработать ошибку
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return base64.StdEncoding.EncodeToString(ciphertext)
}
