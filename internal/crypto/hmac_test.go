package crypto

import (
	"strings"
	"testing"
)

func TestHmacSha1(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		secret   string
		expected string
	}{
		{
			name:     "simple test",
			data:     "test data",
			secret:   "secret",
			expected: "3a5e3a3a3a5e5a5e5a5e5a5e5a5e5a5e5a5e5a5e",
		},
		{
			name:     "empty data",
			data:     "",
			secret:   "secret",
			expected: "",
		},
		{
			name:     "empty secret",
			data:     "data",
			secret:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HmacSha1(tt.data, tt.secret)
			if len(result) != 40 {
				t.Errorf("HmacSha1() length = %d, want 40", len(result))
			}
		})
	}
}

func TestHmacSha1Upper(t *testing.T) {
	data := "test data"
	secret := "secret"

	result := HmacSha1Upper(data, secret)

	if result != strings.ToUpper(result) {
		t.Error("HmacSha1Upper() should return uppercase string")
	}

	if strings.Contains(result, " ") || strings.Contains(result, "-") {
		t.Error("HmacSha1Upper() should not contain spaces or hyphens")
	}
}

func TestHmacSha1Consistency(t *testing.T) {
	data := "consistent data"
	secret := "my_secret_key"

	result1 := HmacSha1(data, secret)
	result2 := HmacSha1(data, secret)

	if result1 != result2 {
		t.Error("HmacSha1() should produce consistent results for same input")
	}
}

func TestSignatureOfHmac(t *testing.T) {
	tests := []struct {
		name          string
		sessionSecret string
		sessionKey    string
		operate       string
		url           string
		date          string
	}{
		{
			name:          "GET request",
			sessionSecret: "secret123",
			sessionKey:    "key456",
			operate:       "GET",
			url:           "https://api.cloud.189.cn/listFiles.action",
			date:          "Mon, 01 Jan 2024 00:00:00 GMT",
		},
		{
			name:          "POST request",
			sessionSecret: "secret123",
			sessionKey:    "key456",
			operate:       "POST",
			url:           "https://api.cloud.189.cn/createFolder.action",
			date:          "Mon, 01 Jan 2024 00:00:00 GMT",
		},
		{
			name:          "URL with query params",
			sessionSecret: "secret123",
			sessionKey:    "key456",
			operate:       "GET",
			url:           "https://api.cloud.189.cn/listFiles.action?folderId=-11&pageNum=1",
			date:          "Mon, 01 Jan 2024 00:00:00 GMT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignatureOfHmac(tt.sessionSecret, tt.sessionKey, tt.operate, tt.url, tt.date)

			if result == "" {
				t.Error("SignatureOfHmac() returned empty string")
			}

			if result != strings.ToUpper(result) {
				t.Error("SignatureOfHmac() should return uppercase string")
			}

			if len(result) != 40 {
				t.Errorf("SignatureOfHmac() length = %d, want 40", len(result))
			}
		})
	}
}

func TestSignatureOfHmacWithParams(t *testing.T) {
	sessionSecret := "secret123"
	sessionKey := "key456"
	operate := "GET"
	url := "https://api.cloud.189.cn/listFiles.action"
	date := "Mon, 01 Jan 2024 00:00:00 GMT"
	params := "encrypted_params_string"

	result := SignatureOfHmacWithParams(sessionSecret, sessionKey, operate, url, date, params)

	if result == "" {
		t.Error("SignatureOfHmacWithParams() returned empty string")
	}

	if result != strings.ToUpper(result) {
		t.Error("SignatureOfHmacWithParams() should return uppercase string")
	}
}

func TestExtractURLPath(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple URL",
			url:      "https://api.cloud.189.cn/listFiles.action",
			expected: "/listFiles.action",
		},
		{
			name:     "URL with query params",
			url:      "https://api.cloud.189.cn/listFiles.action?folderId=-11",
			expected: "/listFiles.action",
		},
		{
			name:     "URL with path segments",
			url:      "https://api.cloud.189.cn/family/file/listFiles.action",
			expected: "/family/file/listFiles.action",
		},
		{
			name:     "URL with fragment",
			url:      "https://api.cloud.189.cn/path#fragment",
			expected: "/path",
		},
		{
			name:     "root URL",
			url:      "https://api.cloud.189.cn",
			expected: "",
		},
		{
			name:     "invalid URL",
			url:      "not a url",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURLPath(tt.url)
			if result != tt.expected {
				t.Errorf("extractURLPath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAppKeySignatureOfHmac(t *testing.T) {
	appSignatureSecret := "app_secret"
	appKey := "app_key"
	operate := "GET"
	url := "https://api.cloud.189.cn/some/action"
	timestamp := int64(1704067200)

	result := AppKeySignatureOfHmac(appSignatureSecret, appKey, operate, url, timestamp)

	if result == "" {
		t.Error("AppKeySignatureOfHmac() returned empty string")
	}

	if result != strings.ToUpper(result) {
		t.Error("AppKeySignatureOfHmac() should return uppercase string")
	}

	if len(result) != 40 {
		t.Errorf("AppKeySignatureOfHmac() length = %d, want 40", len(result))
	}
}
