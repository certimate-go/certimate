package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AESCryptor interface {
	CBCEncrypt(plaintext []byte) ([]byte, error)
	CBCDecrypt(ciphertext []byte) ([]byte, error)
	GCMEncrypt(plaintext, iv, aad []byte) ([]byte, error)
	GCMDecrypt(ciphertext, iv, aad []byte) ([]byte, error)
}

type aesCryptor struct {
	key []byte
}

func (c *aesCryptor) CBCEncrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	paddedData := c.pkcs7Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(paddedData))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedData)

	return ciphertext, nil
}

func (c *aesCryptor) CBCDecrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return c.pkcs7Unpadding(ciphertext), nil
}

func (c *aesCryptor) GCMEncrypt(plaintext, iv, aad []byte) ([]byte, error) {
	const gcmStandardNonceSize int = 12
	if len(iv) != gcmStandardNonceSize {
		return nil, fmt.Errorf("the iv length must be %d bytes for GCM", gcmStandardNonceSize)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, iv, plaintext, aad)
	return ciphertext, nil
}

func (c *aesCryptor) GCMDecrypt(ciphertext, iv, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, iv, ciphertext, aad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (c *aesCryptor) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (c *aesCryptor) pkcs7Unpadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}

	padding := int(data[length-1])
	if padding > length {
		return data
	}

	return data[:length-padding]
}

func NewAESCryptor(key []byte) AESCryptor {
	return &aesCryptor{key: key}
}
