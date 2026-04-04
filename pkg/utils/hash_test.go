package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestMD5Hash(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "simple string",
			data: []byte("hello"),
		},
		{
			name: "another string",
			data: []byte("test"),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MD5Hash(tt.data)
			if len(result) != 32 {
				t.Errorf("MD5Hash() length = %d, want 32", len(result))
			}
		})
	}
}

func TestMD5HashUpper(t *testing.T) {
	data := []byte("test")

	result := MD5HashUpper(data)

	if result != strings.ToUpper(result) {
		t.Error("MD5HashUpper() should return uppercase string")
	}

	if strings.Contains(result, " ") || strings.Contains(result, "-") {
		t.Error("MD5HashUpper() should not contain spaces or hyphens")
	}
}

func TestMD5HashConsistency(t *testing.T) {
	data := []byte("consistent data")

	result1 := MD5Hash(data)
	result2 := MD5Hash(data)

	if result1 != result2 {
		t.Error("MD5Hash() should produce consistent results")
	}
}

func TestMD5HashDifferentInputs(t *testing.T) {
	data1 := []byte("data1")
	data2 := []byte("data2")

	hash1 := MD5Hash(data1)
	hash2 := MD5Hash(data2)

	if hash1 == hash2 {
		t.Error("Different inputs should produce different hashes")
	}
}

func TestMD5HashReader(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple string",
			data: "hello",
		},
		{
			name: "empty reader",
			data: "",
		},
		{
			name: "multi-byte characters",
			data: "你好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			result, err := MD5HashReader(reader)
			if err != nil {
				t.Fatalf("MD5HashReader() error = %v", err)
			}
			if len(result) != 32 {
				t.Errorf("MD5HashReader() length = %d, want 32", len(result))
			}
		})
	}
}

func TestMD5HashReaderUpper(t *testing.T) {
	reader := strings.NewReader("test")

	result, err := MD5HashReaderUpper(reader)
	if err != nil {
		t.Fatalf("MD5HashReaderUpper() error = %v", err)
	}

	if result != strings.ToUpper(result) {
		t.Error("MD5HashReaderUpper() should return uppercase string")
	}
}

func TestMD5HashReaderLargeData(t *testing.T) {
	largeData := strings.Repeat("a", 1024*1024)
	reader := strings.NewReader(largeData)

	result, err := MD5HashReader(reader)
	if err != nil {
		t.Fatalf("MD5HashReader() error = %v", err)
	}

	if len(result) != 32 {
		t.Errorf("MD5HashReader() length = %d, want 32", len(result))
	}
}

func TestMD5HashAndReaderEquality(t *testing.T) {
	data := []byte("test data")

	hashDirect := MD5Hash(data)

	reader := bytes.NewReader(data)
	hashReader, err := MD5HashReader(reader)
	if err != nil {
		t.Fatalf("MD5HashReader() error = %v", err)
	}

	if hashDirect != hashReader {
		t.Error("MD5Hash() and MD5HashReader() should produce same result")
	}
}
