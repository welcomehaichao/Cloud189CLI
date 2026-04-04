package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.configPath == "" {
		t.Error("configPath should not be empty")
	}
}

func TestManagerLoadEmpty(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ConfigFile)

	manager := &Manager{
		configPath: configPath,
		config:     &Config{},
	}

	err := manager.Load()
	if err != nil {
		t.Errorf("Load() on non-existent file should not error: %v", err)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ConfigFile)

	manager := &Manager{
		configPath: configPath,
		config: &Config{
			Username:      "test@example.com",
			SessionKey:    "test_key",
			SessionSecret: "test_secret",
			ExpiresAt:     time.Now().Add(24 * time.Hour),
		},
	}

	err := manager.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loadManager := &Manager{
		configPath: configPath,
		config:     &Config{},
	}

	err = loadManager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loadManager.config.Username != "test@example.com" {
		t.Errorf("Username = %s, want test@example.com", loadManager.config.Username)
	}
}

func TestManagerGetConfig(t *testing.T) {
	manager := &Manager{
		config: &Config{
			Username: "config_user",
		},
	}

	cfg := manager.GetConfig()
	if cfg.Username != "config_user" {
		t.Errorf("GetConfig().Username = %s, want config_user", cfg.Username)
	}
}

func TestManagerIsLoggedIn(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name: "logged in",
			config: &Config{
				SessionKey:    "key",
				SessionSecret: "secret",
			},
			expected: true,
		},
		{
			name: "missing session key",
			config: &Config{
				SessionKey:    "",
				SessionSecret: "secret",
			},
			expected: false,
		},
		{
			name: "missing session secret",
			config: &Config{
				SessionKey:    "key",
				SessionSecret: "",
			},
			expected: false,
		},
		{
			name:     "empty config",
			config:   &Config{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{config: tt.config}
			result := manager.IsLoggedIn()
			if result != tt.expected {
				t.Errorf("IsLoggedIn() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestManagerNeedRefresh(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expires in 2 hours - no refresh needed",
			expiresAt: time.Now().Add(2 * time.Hour),
			expected:  false,
		},
		{
			name:      "expires in 30 minutes - refresh needed",
			expiresAt: time.Now().Add(30 * time.Minute),
			expected:  true,
		},
		{
			name:      "already expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				config: &Config{
					ExpiresAt: tt.expiresAt,
				},
			}
			result := manager.NeedRefresh()
			if result != tt.expected {
				t.Errorf("NeedRefresh() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestManagerClear(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ConfigFile)

	manager := &Manager{
		configPath: configPath,
		config: &Config{
			Username:      "user_to_clear",
			SessionKey:    "key_to_clear",
			SessionSecret: "secret_to_clear",
		},
	}

	err := manager.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if manager.config.Username != "" {
		t.Error("Username should be cleared")
	}

	if manager.config.SessionKey != "" {
		t.Error("SessionKey should be cleared")
	}
}

func TestManagerSetFamilyID(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ConfigFile)

	manager := &Manager{
		configPath: configPath,
		config:     &Config{},
	}

	err := manager.SetFamilyID("family-123")
	if err != nil {
		t.Fatalf("SetFamilyID() error = %v", err)
	}

	if manager.config.FamilyID != "family-123" {
		t.Errorf("FamilyID = %s, want family-123", manager.config.FamilyID)
	}
}

func TestConfigFields(t *testing.T) {
	now := time.Now()
	config := &Config{
		Version:             2,
		Username:            "test@example.com",
		RefreshToken:        "refresh_token",
		AccessToken:         "access_token",
		SessionKey:          "session_key",
		SessionSecret:       "session_secret",
		FamilySessionKey:    "family_key",
		FamilySessionSecret: "family_secret",
		FamilyID:            "family_id",
		ExpiresAt:           now,
		LastUpdate:          now,
		LogDir:              "/var/log/cloud189",
		LogRetentionDays:    30,
	}

	if config.Version != 2 {
		t.Errorf("Version = %d, want 2", config.Version)
	}

	if config.Username != "test@example.com" {
		t.Errorf("Username = %s", config.Username)
	}

	if config.LogRetentionDays != 30 {
		t.Errorf("LogRetentionDays = %d, want 30", config.LogRetentionDays)
	}
}

func TestEncryptedPrefix(t *testing.T) {
	if EncryptedPrefix != "enc:" {
		t.Errorf("EncryptedPrefix = %s, want enc:", EncryptedPrefix)
	}
}

func TestConfigVersion(t *testing.T) {
	if ConfigVersion != 2 {
		t.Errorf("ConfigVersion = %d, want 2", ConfigVersion)
	}
}

func TestManagerSaveFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ConfigFile)

	manager := &Manager{
		configPath: configPath,
		config: &Config{
			Username: "permission_test",
		},
	}

	err := manager.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if info.Mode().Perm()&0600 != 0600 {
		t.Errorf("File should have read/write permissions for owner")
	}
}
