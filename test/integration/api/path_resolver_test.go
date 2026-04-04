//go:build integration
// +build integration

package api_test

import (
	"testing"

	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/internal/config"
)

func TestIntegrationPathResolverRoot(t *testing.T) {
	client := getTestClient(t)

	resolver := api.NewPathResolver(client, false)

	folderID, err := resolver.ResolvePath("/")
	if err != nil {
		t.Fatalf("ResolvePath() error = %v", err)
	}

	if folderID != "-11" {
		t.Errorf("Root folder ID = %s, want -11", folderID)
	}
}

func TestIntegrationPathResolverCache(t *testing.T) {
	client := getTestClient(t)

	resolver := api.NewPathResolver(client, false)

	folderID1, err := resolver.ResolvePath("/")
	if err != nil {
		t.Fatalf("First ResolvePath() error = %v", err)
	}

	folderID2, err := resolver.ResolvePath("/")
	if err != nil {
		t.Fatalf("Second ResolvePath() error = %v", err)
	}

	if folderID1 != folderID2 {
		t.Error("Cached result should match")
	}
}

func TestIntegrationPathResolverInvalidPath(t *testing.T) {
	client := getTestClient(t)

	resolver := api.NewPathResolver(client, false)

	_, err := resolver.ResolvePath("/nonexistent_folder_12345")
	if err == nil {
		t.Error("ResolvePath() should fail for non-existent folder")
	}
}

func TestIntegrationPathResolverFamilyRoot(t *testing.T) {
	client := getTestClient(t)

	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	manager.Load()
	cfg := manager.GetConfig()

	familyID, err := client.GetFamilyID()
	if err != nil {
		t.Skip("No family cloud available")
	}

	cfg.FamilyID = familyID

	resolver := api.NewPathResolver(client, true)

	folderID, err := resolver.ResolvePath("/")
	if err != nil {
		t.Fatalf("ResolvePath() family error = %v", err)
	}

	t.Logf("Family root folder ID: %s", folderID)
}

func TestIntegrationPathResolverGetParentPath(t *testing.T) {
	client := getTestClient(t)

	resolver := api.NewPathResolver(client, false)

	dir, base := resolver.GetParentPath("/folder/file.txt")

	if dir != "/folder" {
		t.Errorf("Dir = %s, want /folder", dir)
	}

	if base != "file.txt" {
		t.Errorf("Base = %s, want file.txt", base)
	}
}
