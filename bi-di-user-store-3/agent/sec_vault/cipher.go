package sec_vault

import (
	"bi-di-user-store-3/agent/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	SECRET_KEY_PREFIX   = "$secret{"
	SECRET_KEY_SUFFIX   = "}"
	SECRET_VALUE_PREFIX = "["
	SECRET_VALUE_SUFFIX = "]"
)

// Encrypt plaintext using AES-GCM.
func Encrypt(plainText string, key []byte) (string, error) {

	secretValue := strings.TrimPrefix(plainText, SECRET_VALUE_PREFIX)
	secretValue = strings.TrimSuffix(secretValue, SECRET_VALUE_SUFFIX)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(secretValue), nil)

	return hex.EncodeToString(cipherText), nil
}

// Decrypt ciphertext using AES-GCM.
func decrypt(cipherText string, key []byte) (string, error) {

	data, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("cipher text is too short")
	}

	nonce, encryptedText := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func ResolveSecret(configValue string, config config.Config, encKey string) (string, error) {

	secretKey := strings.TrimPrefix(configValue, SECRET_KEY_PREFIX)
	secretKey = strings.TrimSuffix(secretKey, SECRET_KEY_SUFFIX)

	encryptedText := config.Secrets[secretKey]

	if encryptedText == "" {
		return "", fmt.Errorf("secret not found")
	}

	decrypted, err := decrypt(encryptedText, []byte(encKey))
	if err != nil {
		return "", err
	}

	return decrypted, nil
}
