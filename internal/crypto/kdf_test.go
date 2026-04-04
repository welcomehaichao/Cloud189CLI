package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key1, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if len(key1) != KeyLength {
		t.Errorf("DeriveKey() length = %d, want %d", len(key1), KeyLength)
	}

	key2, err := DeriveKey()
	if err != nil {
		t.Fatalf("Second DeriveKey() error = %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey() should produce consistent keys on same machine")
	}
}

func TestDeriveKeyWithPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "my_secure_password",
			wantError: false,
		},
		{
			name:      "simple password",
			password:  "123",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
		},
		{
			name:      "long password",
			password:  "this_is_a_very_long_password_string_for_testing",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := DeriveKeyWithPassword(tt.password)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveKeyWithPassword() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if len(key) != KeyLength {
					t.Errorf("DeriveKeyWithPassword() length = %d, want %d", len(key), KeyLength)
				}
			}
		})
	}
}

func TestDeriveKeyWithPasswordConsistency(t *testing.T) {
	password := "test_password_consistency"

	key1, err := DeriveKeyWithPassword(password)
	if err != nil {
		t.Fatalf("First DeriveKeyWithPassword() error = %v", err)
	}

	key2, err := DeriveKeyWithPassword(password)
	if err != nil {
		t.Fatalf("Second DeriveKeyWithPassword() error = %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKeyWithPassword() should produce consistent keys for same password")
	}
}

func TestDeriveKeyWithPasswordUniqueness(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	key1, err := DeriveKeyWithPassword(password1)
	if err != nil {
		t.Fatalf("DeriveKeyWithPassword(password1) error = %v", err)
	}

	key2, err := DeriveKeyWithPassword(password2)
	if err != nil {
		t.Fatalf("DeriveKeyWithPassword(password2) error = %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("Different passwords should produce different keys")
	}
}

func TestHashForComparison(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple string",
			data: "test data",
		},
		{
			name: "empty string",
			data: "",
		},
		{
			name: "long string",
			data: "this is a longer string for testing the hash comparison function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashForComparison(tt.data)

			if len(hash) != 64 {
				t.Errorf("HashForComparison() length = %d, want 64", len(hash))
			}

			if !isHexString(hash) {
				t.Errorf("HashForComparison() result is not valid hex: %s", hash)
			}
		})
	}
}

func TestVerifyHash(t *testing.T) {
	data := "test data for verification"

	hash := HashForComparison(data)

	if !VerifyHash(data, hash) {
		t.Error("VerifyHash() should return true for correct hash")
	}

	if VerifyHash(data, "invalid_hash") {
		t.Error("VerifyHash() should return false for incorrect hash")
	}

	if VerifyHash("different data", hash) {
		t.Error("VerifyHash() should return false for different data")
	}
}

func TestGetMachineID(t *testing.T) {
	id1, err := getMachineID()
	if err != nil {
		t.Fatalf("getMachineID() error = %v", err)
	}

	if id1 == "" {
		t.Error("getMachineID() returned empty string")
	}

	id2, err := getMachineID()
	if err != nil {
		t.Fatalf("Second getMachineID() error = %v", err)
	}

	if id1 != id2 {
		t.Error("getMachineID() should return consistent values")
	}
}

func TestDeriveKeyIntegrationWithGCM(t *testing.T) {
	key, err := DeriveKey()
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	plaintext := []byte("Test message for integration")

	encrypted, err := EncryptGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptGCM() error = %v", err)
	}

	decrypted, err := DecryptGCM(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptGCM() error = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Integration test failed: decrypted data doesn't match original")
	}
}
