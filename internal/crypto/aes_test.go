package crypto

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestAesEncrypt(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		key       []byte
		wantError bool
	}{
		{
			name:      "valid 16-byte key",
			data:      []byte("test data for encryption"),
			key:       []byte("1234567890123456"),
			wantError: false,
		},
		{
			name:      "valid 24-byte key",
			data:      []byte("test data"),
			key:       []byte("123456789012345678901234"),
			wantError: false,
		},
		{
			name:      "valid 32-byte key",
			data:      []byte("test data"),
			key:       []byte("12345678901234567890123456789012"),
			wantError: false,
		},
		{
			name:      "empty data",
			data:      []byte{},
			key:       []byte("1234567890123456"),
			wantError: false,
		},
		{
			name:      "invalid key length",
			data:      []byte("test"),
			key:       []byte("short"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := AesEncrypt(tt.data, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("AesEncrypt() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(encrypted) == 0 && len(tt.data) > 0 {
				t.Error("AesEncrypt() returned empty result")
			}
		})
	}
}

func TestAesEncryptDecrypt(t *testing.T) {
	key := []byte("1234567890123456")
	testData := []byte("Hello, this is a test message for AES encryption!")

	encrypted, err := AesEncrypt(testData, key)
	if err != nil {
		t.Fatalf("AesEncrypt() error = %v", err)
	}

	if bytes.Equal(testData, encrypted) {
		t.Error("Encrypted data should not equal original data")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("Failed to create cipher: %v", err)
	}

	decrypted := make([]byte, len(encrypted))
	size := block.BlockSize()
	for bs, be := 0, size; bs < len(encrypted); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	decrypted = PKCS7UnPadding(decrypted)
	if !bytes.Equal(testData, decrypted) {
		t.Errorf("Decrypted data does not match original. Got %s, want %s", decrypted, testData)
	}
}

func TestAesEncryptHex(t *testing.T) {
	key := []byte("1234567890123456")
	testData := []byte("test hex encoding")

	hexResult, err := AesEncryptHex(testData, key)
	if err != nil {
		t.Fatalf("AesEncryptHex() error = %v", err)
	}

	if !isHexString(hexResult) {
		t.Errorf("AesEncryptHex() result is not valid hex: %s", hexResult)
	}

	if hexResult != strings.ToUpper(hexResult) {
		t.Error("AesEncryptHex() should return uppercase hex string")
	}
}

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
	}{
		{
			name:      "short data",
			data:      []byte("short"),
			blockSize: aes.BlockSize,
		},
		{
			name:      "exact block size",
			data:      make([]byte, aes.BlockSize),
			blockSize: aes.BlockSize,
		},
		{
			name:      "multiple blocks + 1",
			data:      make([]byte, aes.BlockSize+1),
			blockSize: aes.BlockSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := PKCS7Padding(tt.data, tt.blockSize)
			if len(padded)%tt.blockSize != 0 {
				t.Errorf("PKCS7Padding() result length %d is not multiple of block size %d", len(padded), tt.blockSize)
			}
		})
	}
}

func TestEncryptGCM(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		wantError bool
	}{
		{
			name:      "valid encryption",
			plaintext: []byte("test message"),
			key:       make([]byte, 32),
			wantError: false,
		},
		{
			name:      "empty plaintext",
			plaintext: []byte{},
			key:       make([]byte, 32),
			wantError: false,
		},
		{
			name:      "invalid key length",
			plaintext: []byte("test"),
			key:       make([]byte, 16),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncryptGCM(tt.plaintext, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("EncryptGCM() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if _, err := base64.StdEncoding.DecodeString(result); err != nil {
					t.Errorf("EncryptGCM() result is not valid base64: %v", err)
				}
			}
		})
	}
}

func TestDecryptGCM(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("Hello, this is a secret message!")

	encrypted, err := EncryptGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptGCM() error = %v", err)
	}

	decrypted, err := DecryptGCM(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptGCM() error = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("DecryptGCM() = %s, want %s", decrypted, plaintext)
	}
}

func TestDecryptGCMInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext string
		key        []byte
		wantError  bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			key:        make([]byte, 32),
			wantError:  true,
		},
		{
			name:       "invalid key length",
			ciphertext: base64.StdEncoding.EncodeToString([]byte("some data")),
			key:        make([]byte, 16),
			wantError:  true,
		},
		{
			name:       "too short ciphertext",
			ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
			key:        make([]byte, 32),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptGCM(tt.ciphertext, tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("DecryptGCM() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestEncryptDecryptGCMRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	testCases := [][]byte{
		[]byte("short"),
		[]byte("medium length message for testing"),
		make([]byte, 1024),
		make([]byte, 16*1024),
	}

	for i, tc := range testCases {
		t.Run("case_"+string(rune('0'+i)), func(t *testing.T) {
			encrypted, err := EncryptGCM(tc, key)
			if err != nil {
				t.Fatalf("EncryptGCM() error = %v", err)
			}

			decrypted, err := DecryptGCM(encrypted, key)
			if err != nil {
				t.Fatalf("DecryptGCM() error = %v", err)
			}

			if !bytes.Equal(tc, decrypted) {
				t.Error("Round trip encryption/decryption failed")
			}
		})
	}
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func PKCS7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
