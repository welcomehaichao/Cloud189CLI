package api

import (
	"testing"

	"github.com/yuhaichao/cloud189-cli/internal/config"
)

func TestNewPathResolver(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	resolver := NewPathResolver(client, false)

	if resolver == nil {
		t.Fatal("NewPathResolver() returned nil")
	}

	if resolver.client != client {
		t.Error("PathResolver client not set correctly")
	}

	if resolver.isFamily {
		t.Error("PathResolver should not be family by default")
	}
}

func TestNewPathResolverFamily(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	resolver := NewPathResolver(client, true)

	if !resolver.isFamily {
		t.Error("PathResolver should be family")
	}
}

func TestPathResolverResolvePathRoot(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "-11",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "-11",
		},
		{
			name:     "whitespace path",
			path:     "   ",
			expected: "-11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolvePath(tt.path)
			if err != nil {
				t.Errorf("ResolvePath() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("ResolvePath() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestPathResolverResolvePathFamilyRoot(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, true)

	result, err := resolver.ResolvePath("/")
	if err != nil {
		t.Errorf("ResolvePath() error = %v", err)
	}

	if result != "" {
		t.Errorf("Family root path = %s, want empty string", result)
	}
}

func TestPathResolverClearCache(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	resolver.cache["/test"] = "cached_id"

	resolver.ClearCache()

	if len(resolver.cache) != 0 {
		t.Error("ClearCache() should empty the cache")
	}
}

func TestPathResolverGetParentPath(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	tests := []struct {
		name         string
		path         string
		expectedDir  string
		expectedBase string
	}{
		{
			name:         "simple path",
			path:         "/folder/file.txt",
			expectedDir:  "/folder",
			expectedBase: "file.txt",
		},
		{
			name:         "root file",
			path:         "/file.txt",
			expectedDir:  "/",
			expectedBase: "file.txt",
		},
		{
			name:         "single component",
			path:         "file.txt",
			expectedDir:  "-11",
			expectedBase: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, base := resolver.GetParentPath(tt.path)
			if dir != tt.expectedDir {
				t.Errorf("GetParentPath() dir = %s, want %s", dir, tt.expectedDir)
			}
			if base != tt.expectedBase {
				t.Errorf("GetParentPath() base = %s, want %s", base, tt.expectedBase)
			}
		})
	}
}

func TestPathResolverGetParentPathFamily(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, true)

	dir, base := resolver.GetParentPath("file.txt")

	if dir != "" {
		t.Errorf("Family single component dir = %s, want empty", dir)
	}

	if base != "file.txt" {
		t.Errorf("Family base = %s, want file.txt", base)
	}
}

func TestPathResolverCache(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	resolver.cacheMutex.Lock()
	resolver.cache["/cached/path"] = "cached_folder_id"
	resolver.cacheMutex.Unlock()

	result, err := resolver.ResolvePath("/cached/path")
	if err != nil {
		t.Errorf("ResolvePath() with cache error = %v", err)
	}

	if result != "cached_folder_id" {
		t.Errorf("ResolvePath() from cache = %s, want cached_folder_id", result)
	}
}

func TestPathResolverResolvePathWithCreateNotExists(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	_, err := resolver.ResolvePathWithCreate("/nonexistent/path", false)
	if err == nil {
		t.Error("ResolvePathWithCreate() should fail for non-existent path when createIfNotExists is false")
	}
}

func TestPathResolverResolvePathWithCreate(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)
	resolver := NewPathResolver(client, false)

	result, err := resolver.ResolvePathWithCreate("/", false)
	if err != nil {
		t.Errorf("ResolvePathWithCreate() error = %v", err)
	}

	if result != "-11" {
		t.Errorf("ResolvePathWithCreate() = %s, want -11", result)
	}
}

func TestToFamilyOrderBy(t *testing.T) {
	tests := []struct {
		orderBy  string
		expected string
	}{
		{"filename", "1"},
		{"filesize", "2"},
		{"lastOpTime", "3"},
		{"unknown", "1"},
		{"", "1"},
	}

	for _, tt := range tests {
		result := toFamilyOrderBy(tt.orderBy)
		if result != tt.expected {
			t.Errorf("toFamilyOrderBy(%s) = %s, want %s", tt.orderBy, result, tt.expected)
		}
	}
}

func TestToDesc(t *testing.T) {
	tests := []struct {
		orderDirection string
		expected       string
	}{
		{"desc", "true"},
		{"asc", "false"},
		{"", "false"},
		{"invalid", "false"},
	}

	for _, tt := range tests {
		result := toDesc(tt.orderDirection)
		if result != tt.expected {
			t.Errorf("toDesc(%s) = %s, want %s", tt.orderDirection, result, tt.expected)
		}
	}
}
