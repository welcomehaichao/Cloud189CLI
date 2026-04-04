package types

import (
	"testing"
	"time"
)

func TestFileGetID(t *testing.T) {
	file := &File{
		ID:   "12345",
		Name: "test.txt",
	}

	if file.GetID() != "12345" {
		t.Errorf("GetID() = %s, want 12345", file.GetID())
	}
}

func TestFileGetName(t *testing.T) {
	file := &File{
		Name: "document.pdf",
	}

	if file.GetName() != "document.pdf" {
		t.Errorf("GetName() = %s, want document.pdf", file.GetName())
	}
}

func TestFileGetSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected int64
	}{
		{
			name:     "zero size",
			size:     0,
			expected: 0,
		},
		{
			name:     "small file",
			size:     1024,
			expected: 1024,
		},
		{
			name:     "large file",
			size:     1024 * 1024 * 100,
			expected: 104857600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{Size: tt.size}
			if file.GetSize() != tt.expected {
				t.Errorf("GetSize() = %d, want %d", file.GetSize(), tt.expected)
			}
		})
	}
}

func TestFileIsDirectory(t *testing.T) {
	tests := []struct {
		name     string
		isDir    bool
		expected bool
	}{
		{
			name:     "is directory",
			isDir:    true,
			expected: true,
		},
		{
			name:     "is file",
			isDir:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{IsDir: tt.isDir}
			if file.IsDirectory() != tt.expected {
				t.Errorf("IsDirectory() = %v, want %v", file.IsDirectory(), tt.expected)
			}
		})
	}
}

func TestFileGetModified(t *testing.T) {
	now := time.Now()
	file := &File{
		Modified: now,
	}

	if !file.GetModified().Equal(now) {
		t.Errorf("GetModified() = %v, want %v", file.GetModified(), now)
	}
}

func TestFolderGetID(t *testing.T) {
	folder := &Folder{
		ID:   "folder-123",
		Name: "Documents",
	}

	if folder.GetID() != "folder-123" {
		t.Errorf("GetID() = %s, want folder-123", folder.GetID())
	}
}

func TestFolderGetName(t *testing.T) {
	folder := &Folder{
		Name: "MyFolder",
	}

	if folder.GetName() != "MyFolder" {
		t.Errorf("GetName() = %s, want MyFolder", folder.GetName())
	}
}

func TestFolderGetSize(t *testing.T) {
	folder := &Folder{
		Name: "Folder",
	}

	if folder.GetSize() != 0 {
		t.Errorf("Folder GetSize() should always return 0, got %d", folder.GetSize())
	}
}

func TestFolderIsDirectory(t *testing.T) {
	folder := &Folder{}

	if !folder.IsDirectory() {
		t.Error("Folder IsDirectory() should always return true")
	}
}

func TestFolderGetModified(t *testing.T) {
	now := time.Now()
	folder := &Folder{
		Modified: now,
	}

	if !folder.GetModified().Equal(now) {
		t.Errorf("GetModified() = %v, want %v", folder.GetModified(), now)
	}
}

func TestFileWithIcon(t *testing.T) {
	file := &File{
		ID:   "file-1",
		Name: "image.jpg",
		Size: 50000,
		Icon: Icon{
			SmallURL:  "https://example.com/small.jpg",
			LargeURL:  "https://example.com/large.jpg",
			Max600:    "https://example.com/max600.jpg",
			MediumURL: "https://example.com/medium.jpg",
		},
	}

	if file.Icon.SmallURL != "https://example.com/small.jpg" {
		t.Errorf("Icon.SmallURL = %s", file.Icon.SmallURL)
	}

	if file.Icon.LargeURL != "https://example.com/large.jpg" {
		t.Errorf("Icon.LargeURL = %s", file.Icon.LargeURL)
	}
}

func TestFileComplete(t *testing.T) {
	now := time.Now()
	file := &File{
		ID:       "complete-file",
		Name:     "complete.txt",
		Size:     1000,
		IsDir:    false,
		MD5:      "abc123def456",
		Modified: now,
		Created:  now.Add(-24 * time.Hour),
		ParentID: 11,
	}

	if file.ID != "complete-file" {
		t.Error("File ID mismatch")
	}
	if file.Name != "complete.txt" {
		t.Error("File Name mismatch")
	}
	if file.Size != 1000 {
		t.Error("File Size mismatch")
	}
	if file.IsDir {
		t.Error("File should not be directory")
	}
	if file.MD5 != "abc123def456" {
		t.Error("File MD5 mismatch")
	}
}
