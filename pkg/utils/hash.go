package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"strings"
)

func MD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func MD5HashUpper(data []byte) string {
	return strings.ToUpper(MD5Hash(data))
}

func MD5HashReader(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func MD5HashReaderUpper(reader io.Reader) (string, error) {
	hash, err := MD5HashReader(reader)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hash), nil
}
