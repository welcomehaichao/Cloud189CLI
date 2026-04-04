package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewAuditLogger(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewAuditLogger(tempDir, 30, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}

	if logger == nil {
		t.Fatal("NewAuditLogger() returned nil")
	}

	if logger.logDir != tempDir {
		t.Errorf("logDir = %s, want %s", logger.logDir, tempDir)
	}

	defer logger.Close()
}

func TestNewAuditLoggerDefaultDir(t *testing.T) {
	logger, err := NewAuditLogger("", 0, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}

	if logger.logDir == "" {
		t.Error("logDir should have default value")
	}

	if logger.retentionDays != DefaultRetentionDays {
		t.Errorf("retentionDays = %d, want %d", logger.retentionDays, DefaultRetentionDays)
	}

	defer logger.Close()
}

func TestAuditLoggerLog(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewAuditLogger(tempDir, 30, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}
	defer logger.Close()

	entry := &LogEntry{
		Timestamp: time.Now(),
		Username:  "test_user",
		Action:    "login",
		Target:    "system",
		Result:    "success",
		Duration:  time.Second,
	}

	err = logger.Log(entry)
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+LogFileSuffix)
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file should be created")
	}
}

func TestAuditLoggerLogSimple(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewAuditLogger(tempDir, 30, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}
	defer logger.Close()

	err = logger.LogSimple("upload", "test.txt", "success", time.Second)
	if err != nil {
		t.Fatalf("LogSimple() error = %v", err)
	}
}

func TestAuditLoggerLogWithError(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewAuditLogger(tempDir, 30, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}
	defer logger.Close()

	err = logger.LogWithError("download", "file.txt", "failed", time.Second, "network error")
	if err != nil {
		t.Fatalf("LogWithError() error = %v", err)
	}
}

func TestAuditLoggerMaskSensitive(t *testing.T) {
	logger := &AuditLogger{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "email address",
			input:    "test@example.com",
			expected: "tes***@example.com",
		},
		{
			name:     "short email (name <= 3 chars)",
			input:    "a@b.com",
			expected: "a@b.com",
		},
		{
			name:     "long string",
			input:    "12345678abcdefgh",
			expected: "1234***efgh",
		},
		{
			name:     "short string",
			input:    "abc",
			expected: "***",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.maskSensitive(tt.input)
			if result != tt.expected {
				t.Errorf("maskSensitive() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestAuditLoggerFormatEntry(t *testing.T) {
	logger := &AuditLogger{currentUser: "test_user"}

	entry := &LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Username:  "test@example.com",
		Action:    "upload",
		Target:    "/path/to/file.txt",
		FileSize:  1024,
		Result:    "success",
		Duration:  500 * time.Millisecond,
	}

	result := logger.formatEntry(entry)

	if !strings.Contains(result, "upload") {
		t.Error("Formatted entry should contain action")
	}

	if !strings.Contains(result, "success") {
		t.Error("Formatted entry should contain result")
	}

	if !strings.Contains(result, "tes***@example.com") {
		t.Error("Email should be masked")
	}
}

func TestAuditLoggerFormatEntryWithError(t *testing.T) {
	logger := &AuditLogger{currentUser: "test_user"}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Username:  "test_user",
		Action:    "download",
		Target:    "file.txt",
		Result:    "failed",
		Duration:  time.Second,
		Error:     "connection timeout",
	}

	result := logger.formatEntry(entry)

	if !strings.Contains(result, "ERROR: connection timeout") {
		t.Error("Formatted entry should contain error message")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    500,
			expected: "500B",
		},
		{
			name:     "kilobytes",
			bytes:    2048,
			expected: "2.00KB",
		},
		{
			name:     "megabytes",
			bytes:    1048576,
			expected: "1.00MB",
		},
		{
			name:     "gigabytes",
			bytes:    1073741824,
			expected: "1.00GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatFileSize() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestViewLogs(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, time.Now().Format("2006-01-02")+LogFileSuffix)

	logContent := "2024-01-15 10:00:00 | user | action | target | - | success | 100ms\n"
	logContent += "2024-01-15 10:01:00 | user | action2 | target2 | - | success | 200ms\n"

	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	lines, err := ViewLogs(tempDir, "", 0)
	if err != nil {
		t.Fatalf("ViewLogs() error = %v", err)
	}

	if len(lines) != 2 {
		t.Errorf("ViewLogs() returned %d lines, want 2", len(lines))
	}
}

func TestGetLogStats(t *testing.T) {
	tempDir := t.TempDir()

	logFile := filepath.Join(tempDir, "2024-01-15"+LogFileSuffix)
	if err := os.WriteFile(logFile, []byte("test log content"), 0644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	stats, err := GetLogStats(tempDir)
	if err != nil {
		t.Fatalf("GetLogStats() error = %v", err)
	}

	if stats["total_files"].(int) != 1 {
		t.Errorf("total_files = %d, want 1", stats["total_files"])
	}
}

func TestAuditLoggerClose(t *testing.T) {
	tempDir := t.TempDir()
	logger, err := NewAuditLogger(tempDir, 30, "test_user")
	if err != nil {
		t.Fatalf("NewAuditLogger() error = %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestAuditLoggerGetMethods(t *testing.T) {
	logger := &AuditLogger{
		logDir:        "/test/log",
		retentionDays: 60,
	}

	if logger.GetLogDir() != "/test/log" {
		t.Errorf("GetLogDir() = %s, want /test/log", logger.GetLogDir())
	}

	if logger.GetRetentionDays() != 60 {
		t.Errorf("GetRetentionDays() = %d, want 60", logger.GetRetentionDays())
	}
}

func TestDefaultConstants(t *testing.T) {
	if DefaultRetentionDays != 180 {
		t.Errorf("DefaultRetentionDays = %d, want 180", DefaultRetentionDays)
	}

	if LogFilePrefix != "cloud189-audit" {
		t.Errorf("LogFilePrefix = %s, want cloud189-audit", LogFilePrefix)
	}

	if LogFileSuffix != ".log" {
		t.Errorf("LogFileSuffix = %s, want .log", LogFileSuffix)
	}
}
