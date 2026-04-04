package commands

import (
	"testing"
)

func TestSetVersionInfo(t *testing.T) {
	SetVersionInfo("v1.0.0", "2024-01-01")

	if version != "v1.0.0" {
		t.Errorf("version = %s, want v1.0.0", version)
	}

	if buildTime != "2024-01-01" {
		t.Errorf("buildTime = %s, want 2024-01-01", buildTime)
	}
}

func TestPathResolverGetParent(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedDir  string
		expectedName string
	}{
		{
			name:         "simple path",
			path:         "/folder/file.txt",
			expectedDir:  "/folder",
			expectedName: "file.txt",
		},
		{
			name:         "root file",
			path:         "/file.txt",
			expectedDir:  "/",
			expectedName: "file.txt",
		},
		{
			name:         "nested path",
			path:         "/a/b/c/file.txt",
			expectedDir:  "/a/b/c",
			expectedName: "file.txt",
		},
		{
			name:         "single component",
			path:         "file.txt",
			expectedDir:  "/",
			expectedName: "file.txt",
		},
		{
			name:         "trailing slash",
			path:         "/folder/file.txt/",
			expectedDir:  "/folder",
			expectedName: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, name := pathResolverGetParent(tt.path)
			if dir != tt.expectedDir {
				t.Errorf("pathResolverGetParent() dir = %s, want %s", dir, tt.expectedDir)
			}
			if name != tt.expectedName {
				t.Errorf("pathResolverGetParent() name = %s, want %s", name, tt.expectedName)
			}
		})
	}
}

func TestParseOrderBy(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"filename", "filename"},
		{"name", "filename"},
		{"filesize", "filesize"},
		{"size", "filesize"},
		{"lastOpTime", "lastOpTime"},
		{"time", "lastOpTime"},
		{"unknown", "filename"},
	}

	for _, tt := range tests {
		result := parseOrderBy(tt.input)
		if result != tt.expected {
			t.Errorf("parseOrderBy(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestParseDesc(t *testing.T) {
	tests := []struct {
		desc     bool
		expected string
	}{
		{true, "desc"},
		{false, "asc"},
	}

	for _, tt := range tests {
		result := parseDesc(tt.desc)
		if result != tt.expected {
			t.Errorf("parseDesc(%v) = %s, want %s", tt.desc, result, tt.expected)
		}
	}
}
