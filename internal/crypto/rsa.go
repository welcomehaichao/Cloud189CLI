package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"strings"
)

func RsaEncrypt(publicKey, data string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", nil
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	pub := pubInterface.(*rsa.PublicKey)
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(data))
	if err != nil {
		return "", err
	}

	return strings.ToUpper(hex.EncodeToString(encrypted)), nil
}

func RsaEncryptWithPrefix(publicKey, prefix, data string) (string, error) {
	encrypted, err := RsaEncrypt(publicKey, data)
	if err != nil {
		return "", err
	}
	return prefix + encrypted, nil
}
