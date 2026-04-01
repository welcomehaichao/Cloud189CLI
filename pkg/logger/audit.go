package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	DefaultRetentionDays = 180
	LogFilePrefix        = "cloud189-audit"
	LogFileSuffix        = ".log"
)

type AuditLogger struct {
	logDir        string
	retentionDays int
	currentUser   string
	currentFile   *os.File
	currentDate   string
	mu            sync.Mutex
}

type LogEntry struct {
	Timestamp time.Time
	Username  string
	Action    string
	Target    string
	FileSize  int64
	Result    string
	Duration  time.Duration
	Error     string
}

func NewAuditLogger(logDir string, retentionDays int, username string) (*AuditLogger, error) {
	if logDir == "" {
		logDir = getDefaultLogDir()
	}

	if retentionDays <= 0 {
		retentionDays = DefaultRetentionDays
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &AuditLogger{
		logDir:        logDir,
		retentionDays: retentionDays,
		currentUser:   username,
	}

	if err := logger.cleanOldLogs(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to clean old logs: %v\n", err)
	}

	return logger, nil
}

func (l *AuditLogger) Log(entry *LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.rotateIfNeeded(); err != nil {
		return err
	}

	line := l.formatEntry(entry)
	if _, err := l.currentFile.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

func (l *AuditLogger) LogSimple(action, target, result string, duration time.Duration) error {
	return l.Log(&LogEntry{
		Timestamp: time.Now(),
		Username:  l.currentUser,
		Action:    action,
		Target:    target,
		Result:    result,
		Duration:  duration,
	})
}

func (l *AuditLogger) LogWithError(action, target, result string, duration time.Duration, errMsg string) error {
	return l.Log(&LogEntry{
		Timestamp: time.Now(),
		Username:  l.currentUser,
		Action:    action,
		Target:    target,
		Result:    result,
		Duration:  duration,
		Error:     errMsg,
	})
}

func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentFile != nil {
		return l.currentFile.Close()
	}
	return nil
}

func (l *AuditLogger) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")

	if l.currentDate == today && l.currentFile != nil {
		return nil
	}

	if l.currentFile != nil {
		l.currentFile.Close()
	}

	logPath := filepath.Join(l.logDir, today+LogFileSuffix)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.currentFile = file
	l.currentDate = today

	return nil
}

func (l *AuditLogger) formatEntry(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	username := l.maskSensitive(entry.Username)
	duration := entry.Duration.Round(time.Millisecond)

	var sizeStr string
	if entry.FileSize > 0 {
		sizeStr = formatFileSize(entry.FileSize)
	} else {
		sizeStr = "-"
	}

	line := fmt.Sprintf("%s | %s | %s | %s | %s | %s | %v",
		timestamp,
		username,
		entry.Action,
		entry.Target,
		sizeStr,
		entry.Result,
		duration,
	)

	if entry.Error != "" {
		line += fmt.Sprintf(" | ERROR: %s", entry.Error)
	}

	return line
}

func (l *AuditLogger) maskSensitive(s string) string {
	if s == "" {
		return "-"
	}

	if strings.Contains(s, "@") {
		parts := strings.Split(s, "@")
		if len(parts) == 2 {
			name := parts[0]
			domain := parts[1]
			if len(name) > 3 {
				name = name[:3] + "***"
			}
			return name + "@" + domain
		}
	}

	if len(s) > 8 {
		return s[:4] + "***" + s[len(s)-4:]
	}

	return "***"
}

func (l *AuditLogger) cleanOldLogs() error {
	files, err := os.ReadDir(l.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -l.retentionDays)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), LogFileSuffix) {
			dateStr := strings.TrimSuffix(file.Name(), LogFileSuffix)
			fileDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}

			if fileDate.Before(cutoff) {
				filePath := filepath.Join(l.logDir, file.Name())
				if err := os.Remove(filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove old log %s: %v\n", file.Name(), err)
				}
			}
		}
	}

	return nil
}

func (l *AuditLogger) GetLogDir() string {
	return l.logDir
}

func (l *AuditLogger) GetRetentionDays() int {
	return l.retentionDays
}

func getDefaultLogDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return filepath.Join(homeDir, ".cloud189", "logs")
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func ViewLogs(logDir string, date string, lines int) ([]string, error) {
	if logDir == "" {
		logDir = getDefaultLogDir()
	}

	var logFile string
	if date != "" {
		logFile = filepath.Join(logDir, date+LogFileSuffix)
	} else {
		logFile = filepath.Join(logDir, time.Now().Format("2006-01-02")+LogFileSuffix)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	allLines := strings.Split(string(data), "\n")

	if lines <= 0 || lines > len(allLines) {
		lines = len(allLines)
	}

	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}

	result := allLines[start:]

	var cleanResult []string
	for _, line := range result {
		if strings.TrimSpace(line) != "" {
			cleanResult = append(cleanResult, line)
		}
	}

	return cleanResult, nil
}

func GetLogStats(logDir string) (map[string]interface{}, error) {
	if logDir == "" {
		logDir = getDefaultLogDir()
	}

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var totalSize int64
	logFiles := []map[string]interface{}{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), LogFileSuffix) {
			info, err := file.Info()
			if err != nil {
				continue
			}

			totalSize += info.Size()

			logFiles = append(logFiles, map[string]interface{}{
				"name": file.Name(),
				"size": info.Size(),
				"date": strings.TrimSuffix(file.Name(), LogFileSuffix),
			})
		}
	}

	return map[string]interface{}{
		"total_files": len(logFiles),
		"total_size":  totalSize,
		"log_files":   logFiles,
	}, nil
}
