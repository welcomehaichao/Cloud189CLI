package api

import (
	"net/http"
	"testing"

	"github.com/yuhaichao/cloud189-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		SessionKey:    "test_key",
		SessionSecret: "test_secret",
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config != cfg {
		t.Error("Client config not set correctly")
	}
}

func TestNewClientWithManager(t *testing.T) {
	manager := &config.Manager{}

	client := NewClientWithManager(manager)

	if client == nil {
		t.Fatal("NewClientWithManager() returned nil")
	}
}

func TestClientSuffix(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	suffix := client.ClientSuffix()

	if suffix["clientType"] != PC {
		t.Errorf("clientType = %s, want %s", suffix["clientType"], PC)
	}

	if suffix["version"] != Version {
		t.Errorf("version = %s, want %s", suffix["version"], Version)
	}

	if suffix["channelId"] != ChannelID {
		t.Errorf("channelId = %s, want %s", suffix["channelId"], ChannelID)
	}
}

func TestAuthClientSuffix(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	suffix := client.AuthClientSuffix()

	if suffix["clientType"] != ClientType {
		t.Errorf("clientType = %s, want %s", suffix["clientType"], ClientType)
	}
}

func TestEncryptParams(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		params    Params
		isFamily  bool
		wantEmpty bool
	}{
		{
			name: "valid session secret",
			config: &config.Config{
				SessionSecret: "0123456789abcdef",
			},
			params:    Params{"key": "value"},
			isFamily:  false,
			wantEmpty: false,
		},
		{
			name: "empty session secret",
			config: &config.Config{
				SessionSecret: "",
			},
			params:    Params{"key": "value"},
			isFamily:  false,
			wantEmpty: true,
		},
		{
			name: "short session secret",
			config: &config.Config{
				SessionSecret: "short",
			},
			params:    Params{"key": "value"},
			isFamily:  false,
			wantEmpty: true,
		},
		{
			name: "nil params",
			config: &config.Config{
				SessionSecret: "0123456789abcdef",
			},
			params:    nil,
			isFamily:  false,
			wantEmpty: true,
		},
		{
			name: "family session secret",
			config: &config.Config{
				SessionSecret:       "personal_secret",
				FamilySessionSecret: "family_secret_16",
			},
			params:    Params{"key": "value"},
			isFamily:  true,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			result := client.EncryptParams(tt.params, tt.isFamily)

			if tt.wantEmpty && result != "" {
				t.Errorf("EncryptParams() = %s, want empty", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("EncryptParams() returned empty, want non-empty")
			}
		})
	}
}

func TestParamsSet(t *testing.T) {
	p := Params{}
	p.Set("key", "value")

	if p["key"] != "value" {
		t.Error("Params.Set() did not set value correctly")
	}
}

func TestParamsEncode(t *testing.T) {
	tests := []struct {
		name      string
		params    Params
		wantEmpty bool
	}{
		{
			name:      "nil params",
			params:    nil,
			wantEmpty: true,
		},
		{
			name:      "empty params",
			params:    Params{},
			wantEmpty: true,
		},
		{
			name:      "single param",
			params:    Params{"key": "value"},
			wantEmpty: false,
		},
		{
			name:      "multiple params",
			params:    Params{"b": "2", "a": "1"},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.Encode()

			if tt.wantEmpty && result != "" {
				t.Errorf("Encode() = %s, want empty", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("Encode() returned empty, want non-empty")
			}
		})
	}
}

func TestParamsEncodeOrder(t *testing.T) {
	p := Params{
		"z": "3",
		"a": "1",
		"m": "2",
	}

	result := p.Encode()

	if result != "a=1&m=2&z=3" {
		t.Errorf("Encode() = %s, want sorted params", result)
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"AccountType", AccountType, "02"},
		{"AppID", AppID, "8025431004"},
		{"ClientType", ClientType, "10020"},
		{"Version", Version, "6.2"},
		{"PC", PC, "TELEPC"},
		{"ChannelID", ChannelID, "web_cloud.189.cn"},
		{"WebURL", WebURL, "https://cloud.189.cn"},
		{"AuthURL", AuthURL, "https://open.e.189.cn"},
		{"APIURL", APIURL, "https://api.cloud.189.cn"},
		{"UploadURL", UploadURL, "https://upload.cloud.189.cn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %s, want %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestClientGetClient(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	restyClient := client.getClient()

	if restyClient == nil {
		t.Error("getClient() returned nil")
	}
}

func TestSignatureHeaderFields(t *testing.T) {
	cfg := &config.Config{
		SessionKey:    "test_key",
		SessionSecret: "test_secret_16_chars",
	}
	client := NewClient(cfg)

	headers := client.SignatureHeader("https://api.cloud.189.cn/test", http.MethodGet, false)

	requiredHeaders := []string{"Date", "SessionKey", "X-Request-ID", "Signature"}
	for _, h := range requiredHeaders {
		if _, exists := headers[h]; !exists {
			t.Errorf("SignatureHeader missing required header: %s", h)
		}
	}
}

func TestSignatureHeaderWithParams(t *testing.T) {
	cfg := &config.Config{
		SessionKey:    "test_key",
		SessionSecret: "test_secret_16_chars",
	}
	client := NewClient(cfg)

	headers := client.SignatureHeaderWithParams("https://api.cloud.189.cn/test", http.MethodGet, "params_data", false)

	if _, exists := headers["Signature"]; !exists {
		t.Error("SignatureHeaderWithParams missing Signature header")
	}
}

func TestSignatureHeaderFamily(t *testing.T) {
	cfg := &config.Config{
		SessionKey:          "personal_key",
		SessionSecret:       "personal_secret_16",
		FamilySessionKey:    "family_key",
		FamilySessionSecret: "family_secret_16",
	}
	client := NewClient(cfg)

	personalHeaders := client.SignatureHeader("https://api.cloud.189.cn/test", http.MethodGet, false)
	familyHeaders := client.SignatureHeader("https://api.cloud.189.cn/test", http.MethodGet, true)

	if personalHeaders["SessionKey"] == familyHeaders["SessionKey"] {
		t.Error("Family and personal session keys should be different")
	}
}
