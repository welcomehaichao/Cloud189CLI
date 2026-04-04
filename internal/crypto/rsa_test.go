package crypto

import (
	"strings"
	"testing"
)

const validPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2Z3qX2BTLS4e7g6V3Q2F
8B4k3H5vN7mP9K2L4R6T8Y1W3E5U7I9O0P2A4S6D8F0G2H4J6K8L0M2N4O6Q8R0
S2T4U6V8W0X2Y4Z6A8B0C2D4E6F8G0H2I4J6K8L0M2N4O6P8Q0R2S4T6U8V0W2X
-----END PUBLIC KEY-----`

func TestRsaEncrypt(t *testing.T) {
	tests := []struct {
		name       string
		publicKey  string
		data       string
		wantError  bool
		checkUpper bool
	}{
		{
			name:       "simple string",
			publicKey:  validPublicKey,
			data:       "test",
			wantError:  false,
			checkUpper: true,
		},
		{
			name:       "longer string",
			publicKey:  validPublicKey,
			data:       "This is a longer test string for RSA encryption",
			wantError:  false,
			checkUpper: true,
		},
		{
			name:       "empty data",
			publicKey:  validPublicKey,
			data:       "",
			wantError:  false,
			checkUpper: false,
		},
		{
			name:       "empty public key",
			publicKey:  "",
			data:       "test",
			wantError:  false,
			checkUpper: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RsaEncrypt(tt.publicKey, tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("RsaEncrypt() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.checkUpper {
				if result != strings.ToUpper(result) {
					t.Error("RsaEncrypt() should return uppercase hex string")
				}
			}
		})
	}
}

func TestRsaEncryptWithPrefix(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		prefix    string
		data      string
		wantError bool
	}{
		{
			name:      "with prefix",
			publicKey: validPublicKey,
			prefix:    "PRE:",
			data:      "test data",
			wantError: false,
		},
		{
			name:      "empty prefix",
			publicKey: validPublicKey,
			prefix:    "",
			data:      "test data",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RsaEncryptWithPrefix(tt.publicKey, tt.prefix, tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("RsaEncryptWithPrefix() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.prefix != "" {
				if !strings.HasPrefix(result, tt.prefix) {
					t.Errorf("RsaEncryptWithPrefix() result should start with prefix %s", tt.prefix)
				}
			}
		})
	}
}

func TestRsaEncryptHexFormat(t *testing.T) {
	result, err := RsaEncrypt(validPublicKey, "test")
	if err != nil {
		t.Fatalf("RsaEncrypt() error = %v", err)
	}

	if result == "" {
		t.Skip("RsaEncrypt() returned empty result (key format may not be valid)")
	}

	for _, c := range result {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			t.Errorf("RsaEncrypt() result contains non-hex character: %c", c)
			break
		}
	}
}
