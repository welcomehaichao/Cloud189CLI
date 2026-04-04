//go:build integration
// +build integration

package api_test

import (
	"os"
	"testing"

	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/internal/config"
)

func getTestClient(t *testing.T) *api.Client {
	t.Helper()

	configPath := os.Getenv("CLOUD189_TEST_CONFIG")
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = homeDir + "/.cloud189/config.json"
	}

	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !manager.IsLoggedIn() {
		t.Skip("Not logged in. Run 'cloud189 login' first or set CLOUD189_TEST_CONFIG")
	}

	return api.NewClientWithManager(manager)
}

func TestIntegrationPasswordLogin(t *testing.T) {
	username := os.Getenv("CLOUD189_USERNAME")
	password := os.Getenv("CLOUD189_PASSWORD")

	if username == "" || password == "" {
		t.Skip("CLOUD189_USERNAME and CLOUD189_PASSWORD not set")
	}

	cfg := &config.Config{}
	client := api.NewClient(cfg)

	session, err := client.Login(username, password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if session.SessionKey == "" {
		t.Error("Session key should not be empty after login")
	}

	if session.LoginName == "" {
		t.Error("Login name should not be empty after login")
	}

	t.Logf("Login successful: %s", session.LoginName)
}

func TestIntegrationQRCodeLogin(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_QR") == "" {
		t.Skip("Set CLOUD189_TEST_QR=1 to test QR code login")
	}

	cfg := &config.Config{}
	client := api.NewClient(cfg)

	session, err := client.LoginByQRCode()
	if err != nil {
		t.Fatalf("LoginByQRCode() error = %v", err)
	}

	if session.SessionKey == "" {
		t.Error("Session key should not be empty after QR login")
	}

	t.Logf("QR Login successful: %s", session.LoginName)
}

func TestIntegrationTokenRefresh(t *testing.T) {
	client := getTestClient(t)

	cfg := getConfig(t)
	if cfg.RefreshToken == "" {
		t.Skip("No refresh token available")
	}

	session, err := client.RefreshToken(cfg.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if session.SessionKey == "" {
		t.Error("Session key should not be empty after refresh")
	}

	t.Log("Token refresh successful")
}

func getConfig(t *testing.T) *config.Config {
	t.Helper()

	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	return manager.GetConfig()
}

func TestIntegrationKeepAlive(t *testing.T) {
	client := getTestClient(t)

	err := client.KeepAlive()
	if err != nil {
		t.Fatalf("KeepAlive() error = %v", err)
	}

	t.Log("KeepAlive successful")
}

func TestIntegrationGetCapacityInfo(t *testing.T) {
	client := getTestClient(t)

	capacity, err := client.GetCapacityInfo()
	if err != nil {
		t.Fatalf("GetCapacityInfo() error = %v", err)
	}

	if capacity.ResCode != 0 {
		t.Errorf("GetCapacityInfo() res_code = %d, want 0", capacity.ResCode)
	}

	if capacity.CloudCapacityInfo.TotalSize == 0 {
		t.Error("Total size should not be zero")
	}

	t.Logf("Personal cloud: %d/%d bytes used",
		capacity.CloudCapacityInfo.UsedSize,
		capacity.CloudCapacityInfo.TotalSize)
	t.Logf("Family cloud: %d/%d bytes used",
		capacity.FamilyCapacityInfo.UsedSize,
		capacity.FamilyCapacityInfo.TotalSize)
}

func TestIntegrationGetFamilyList(t *testing.T) {
	client := getTestClient(t)

	families, err := client.GetFamilyList()
	if err != nil {
		t.Fatalf("GetFamilyList() error = %v", err)
	}

	t.Logf("Found %d family clouds", len(families))

	for _, f := range families {
		t.Logf("  - Family: %s (ID: %d)", f.RemarkName, f.FamilyID)
	}
}

func TestIntegrationGetFamilyID(t *testing.T) {
	client := getTestClient(t)

	familyID, err := client.GetFamilyID()
	if err != nil {
		t.Logf("GetFamilyID() error = %v (may have no family)", err)
		return
	}

	if familyID == "" {
		t.Error("Family ID should not be empty if family exists")
	}

	t.Logf("Family ID: %s", familyID)
}
