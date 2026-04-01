package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// AesEncrypt ECB模式加密（不推荐使用，仅保留兼容性）
func AesEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data = PKCS7Padding(data, block.BlockSize())
	encrypted := make([]byte, len(data))

	size := block.BlockSize()
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(encrypted[bs:be], data[bs:be])
	}

	return encrypted, nil
}

func AesEncryptHex(data, key []byte) (string, error) {
	encrypted, err := AesEncrypt(data, key)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(encrypted)), nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// EncryptGCM 使用AES-GCM模式加密（推荐）
// 返回格式: base64(nonce + ciphertext + tag)
func EncryptGCM(plaintext []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据（nonce + ciphertext + tag）
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 返回base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptGCM 使用AES-GCM模式解密
func DecryptGCM(ciphertext string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// 提取nonce和实际密文
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
