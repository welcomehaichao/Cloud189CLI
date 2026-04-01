package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuhaichao/cloud189-cli/internal/crypto"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

const (
	ConfigDir  = ".cloud189"
	ConfigFile = "config.json"
	// 配置文件版本
	ConfigVersion = 2
	// 加密字段前缀
	EncryptedPrefix = "enc:"
)

type Config struct {
	Version             int       `json:"version,omitempty"`
	Username            string    `json:"username"`
	RefreshToken        string    `json:"refresh_token"`
	AccessToken         string    `json:"access_token"`
	SessionKey          string    `json:"session_key"`
	SessionSecret       string    `json:"session_secret"`
	FamilySessionKey    string    `json:"family_session_key"`
	FamilySessionSecret string    `json:"family_session_secret"`
	FamilyID            string    `json:"family_id"`
	ExpiresAt           time.Time `json:"expires_at"`
	LastUpdate          time.Time `json:"last_update"`
	LogDir              string    `json:"log_dir,omitempty"`
	LogRetentionDays    int       `json:"log_retention_days,omitempty"`
}

// encryptedConfig 用于JSON序列化的加密配置
type encryptedConfig struct {
	Version             int       `json:"version"`
	Username            string    `json:"username"`
	RefreshToken        string    `json:"refresh_token"`
	AccessToken         string    `json:"access_token"`
	SessionKey          string    `json:"session_key"`
	SessionSecret       string    `json:"session_secret"`
	FamilySessionKey    string    `json:"family_session_key"`
	FamilySessionSecret string    `json:"family_session_secret"`
	FamilyID            string    `json:"family_id"`
	ExpiresAt           time.Time `json:"expires_at"`
	LastUpdate          time.Time `json:"last_update"`
	LogDir              string    `json:"log_dir,omitempty"`
	LogRetentionDays    int       `json:"log_retention_days,omitempty"`
}

type Manager struct {
	configPath string
	config     *Config
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ConfigFile)

	return &Manager{
		configPath: configPath,
		config:     &Config{},
	}, nil
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析JSON
	var rawConfig encryptedConfig
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 检查版本，决定是否需要解密
	if rawConfig.Version >= 2 {
		// 新版本：解密敏感字段
		if err := m.decryptConfig(&rawConfig); err != nil {
			return fmt.Errorf("failed to decrypt config: %w", err)
		}
		// 复制到内部配置
		m.config = &Config{
			Version:             rawConfig.Version,
			Username:            rawConfig.Username,
			RefreshToken:        rawConfig.RefreshToken,
			AccessToken:         rawConfig.AccessToken,
			SessionKey:          rawConfig.SessionKey,
			SessionSecret:       rawConfig.SessionSecret,
			FamilySessionKey:    rawConfig.FamilySessionKey,
			FamilySessionSecret: rawConfig.FamilySessionSecret,
			FamilyID:            rawConfig.FamilyID,
			ExpiresAt:           rawConfig.ExpiresAt,
			LastUpdate:          rawConfig.LastUpdate,
			LogDir:              rawConfig.LogDir,
			LogRetentionDays:    rawConfig.LogRetentionDays,
		}
	} else {
		// 旧版本或无版本：明文，需要迁移
		m.config = &Config{
			Version:             0, // 旧版本标记
			Username:            rawConfig.Username,
			RefreshToken:        rawConfig.RefreshToken,
			AccessToken:         rawConfig.AccessToken,
			SessionKey:          rawConfig.SessionKey,
			SessionSecret:       rawConfig.SessionSecret,
			FamilySessionKey:    rawConfig.FamilySessionKey,
			FamilySessionSecret: rawConfig.FamilySessionSecret,
			FamilyID:            rawConfig.FamilyID,
			ExpiresAt:           rawConfig.ExpiresAt,
			LastUpdate:          rawConfig.LastUpdate,
			LogDir:              rawConfig.LogDir,
			LogRetentionDays:    rawConfig.LogRetentionDays,
		}
		// 同步迁移到加密版本（忽略错误，因为配置已加载成功）
		_ = m.Save()
	}

	return nil
}

func (m *Manager) Save() error {
	m.config.LastUpdate = time.Now()

	// 加密敏感字段
	encryptedConfig, err := m.encryptConfig()
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// 序列化JSON
	data, err := json.MarshalIndent(encryptedConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// encryptConfig 加密配置的敏感字段
func (m *Manager) encryptConfig() (*encryptedConfig, error) {
	// 获取加密密钥
	key, err := crypto.DeriveKey()
	if err != nil {
		return nil, err
	}

	config := &encryptedConfig{
		Version:          ConfigVersion,
		Username:         m.config.Username,
		ExpiresAt:        m.config.ExpiresAt,
		LastUpdate:       m.config.LastUpdate,
		LogDir:           m.config.LogDir,
		LogRetentionDays: m.config.LogRetentionDays,
	}

	// 加密敏感字段
	if m.config.RefreshToken != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.RefreshToken), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt refresh_token: %w", err)
		}
		config.RefreshToken = EncryptedPrefix + encrypted
	}

	if m.config.AccessToken != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.AccessToken), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt access_token: %w", err)
		}
		config.AccessToken = EncryptedPrefix + encrypted
	}

	if m.config.SessionKey != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.SessionKey), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt session_key: %w", err)
		}
		config.SessionKey = EncryptedPrefix + encrypted
	}

	if m.config.SessionSecret != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.SessionSecret), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt session_secret: %w", err)
		}
		config.SessionSecret = EncryptedPrefix + encrypted
	}

	if m.config.FamilySessionKey != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.FamilySessionKey), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt family_session_key: %w", err)
		}
		config.FamilySessionKey = EncryptedPrefix + encrypted
	}

	if m.config.FamilySessionSecret != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.FamilySessionSecret), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt family_session_secret: %w", err)
		}
		config.FamilySessionSecret = EncryptedPrefix + encrypted
	}

	if m.config.FamilyID != "" {
		encrypted, err := crypto.EncryptGCM([]byte(m.config.FamilyID), key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt family_id: %w", err)
		}
		config.FamilyID = EncryptedPrefix + encrypted
	}

	return config, nil
}

// decryptConfig 解密配置的敏感字段
func (m *Manager) decryptConfig(rawConfig *encryptedConfig) error {
	// 获取解密密钥
	key, err := crypto.DeriveKey()
	if err != nil {
		return err
	}

	// 解密敏感字段
	if strings.HasPrefix(rawConfig.RefreshToken, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.RefreshToken[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt refresh_token: %w", err)
		}
		rawConfig.RefreshToken = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.AccessToken, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.AccessToken[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt access_token: %w", err)
		}
		rawConfig.AccessToken = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.SessionKey, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.SessionKey[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt session_key: %w", err)
		}
		rawConfig.SessionKey = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.SessionSecret, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.SessionSecret[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt session_secret: %w", err)
		}
		rawConfig.SessionSecret = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.FamilySessionKey, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.FamilySessionKey[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt family_session_key: %w", err)
		}
		rawConfig.FamilySessionKey = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.FamilySessionSecret, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.FamilySessionSecret[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt family_session_secret: %w", err)
		}
		rawConfig.FamilySessionSecret = string(decrypted)
	}

	if strings.HasPrefix(rawConfig.FamilyID, EncryptedPrefix) {
		decrypted, err := crypto.DecryptGCM(rawConfig.FamilyID[len(EncryptedPrefix):], key)
		if err != nil {
			return fmt.Errorf("failed to decrypt family_id: %w", err)
		}
		rawConfig.FamilyID = string(decrypted)
	}

	return nil
}

func (m *Manager) GetConfig() *Config {
	return m.config
}

func (m *Manager) SetSession(session *types.Session) error {
	m.config.Username = session.LoginName
	m.config.RefreshToken = session.RefreshToken
	m.config.AccessToken = session.AccessToken
	m.config.SessionKey = session.SessionKey
	m.config.SessionSecret = session.SessionSecret
	m.config.FamilySessionKey = session.FamilySessionKey
	m.config.FamilySessionSecret = session.FamilySessionSecret
	m.config.ExpiresAt = time.Now().Add(24 * time.Hour)

	return m.Save()
}

func (m *Manager) SetFamilyID(familyID string) error {
	m.config.FamilyID = familyID
	return m.Save()
}

func (m *Manager) Clear() error {
	m.config = &Config{}
	return m.Save()
}

func (m *Manager) IsLoggedIn() bool {
	return m.config.SessionKey != "" && m.config.SessionSecret != ""
}

func (m *Manager) NeedRefresh() bool {
	return time.Now().After(m.config.ExpiresAt.Add(-1 * time.Hour))
}
