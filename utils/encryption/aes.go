package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func AesEncryptByGCM(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceStr := key[:gcm.NonceSize()]
	nonce := []byte(nonceStr)
	seal := gcm.Seal(nonce, nonce, []byte(data), nil)
	return seal, nil
}

func AesDecryptByGCM(data, key []byte) ([]byte, error) {
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
		return nil, fmt.Errorf("data length < %d", nonceSize)
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	open, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return open, nil
}
