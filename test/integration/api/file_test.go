//go:build integration
// +build integration

package api_test

import (
	"testing"
	"time"

	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/internal/config"
)

func TestIntegrationListFiles(t *testing.T) {
	client := getTestClient(t)

	files, err := client.ListFiles("-11", 1, 20, "filename", "asc", false)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	t.Logf("Found %d items in root directory", len(files))

	for i, f := range files {
		if i >= 5 {
			break
		}
		typeStr := "file"
		if f.IsDir {
			typeStr = "folder"
		}
		t.Logf("  - [%s] %s (ID: %s)", typeStr, f.Name, f.ID)
	}
}

func TestIntegrationListFilesFamily(t *testing.T) {
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

	files, err := client.ListFiles("", 1, 20, "filename", "asc", true)
	if err != nil {
		t.Fatalf("ListFiles() family error = %v", err)
	}

	t.Logf("Found %d items in family root", len(files))
}

func TestIntegrationCreateAndDeleteFolder(t *testing.T) {
	client := getTestClient(t)

	folderName := "test_folder_" + time.Now().Format("20060102_150405")

	folder, err := client.CreateFolder("-11", folderName, false)
	if err != nil {
		t.Fatalf("CreateFolder() error = %v", err)
	}

	if folder.ID == "" {
		t.Fatal("Folder ID should not be empty")
	}

	t.Logf("Created folder: %s (ID: %s)", folder.Name, folder.ID)

	time.Sleep(500 * time.Millisecond)

	files, err := client.ListFiles("-11", 1, 100, "filename", "asc", false)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	var found bool
	for _, f := range files {
		if f.Name == folderName {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created folder not found in list")
	}

	t.Log("CreateFolder test passed")
}

func TestIntegrationRenameFolder(t *testing.T) {
	client := getTestClient(t)

	originalName := "rename_test_" + time.Now().Format("20060102_150405")
	newName := "renamed_" + time.Now().Format("20060102_150405")

	folder, err := client.CreateFolder("-11", originalName, false)
	if err != nil {
		t.Fatalf("CreateFolder() error = %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	err = client.RenameFolder(folder.ID, newName, false)
	if err != nil {
		t.Fatalf("RenameFolder() error = %v", err)
	}

	t.Logf("Renamed folder from %s to %s", originalName, newName)
}

func TestIntegrationGetDownloadURL(t *testing.T) {
	client := getTestClient(t)

	files, err := client.ListFiles("-11", 1, 20, "filename", "asc", false)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	var fileID string
	for _, f := range files {
		if !f.IsDir && f.ID != "" {
			fileID = f.ID
			break
		}
	}

	if fileID == "" {
		t.Skip("No files found in root directory")
	}

	url, err := client.GetDownloadURL(fileID, false)
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v", err)
	}

	if url == "" {
		t.Error("Download URL should not be empty")
	}

	t.Logf("Download URL: %s", url)
}

func TestIntegrationCreateShareLink(t *testing.T) {
	client := getTestClient(t)

	files, err := client.ListFiles("-11", 1, 20, "filename", "asc", false)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	var fileID string
	var fileName string
	for _, f := range files {
		if !f.IsDir && f.ID != "" {
			fileID = f.ID
			fileName = f.Name
			break
		}
	}

	if fileID == "" {
		t.Skip("No files found in root directory")
	}

	share, err := client.CreateShareLink(fileID, false, 7, "", false)
	if err != nil {
		t.Fatalf("CreateShareLink() error = %v", err)
	}

	if share.ShareId == "" {
		t.Error("Share ID should not be empty")
	}

	t.Logf("Created share for %s:", fileName)
	t.Logf("  Share ID: %s", share.ShareId)
	t.Logf("  Share Link: %s", share.ShareLink)
	t.Logf("  Access Code: %s", share.AccessCode)
}
