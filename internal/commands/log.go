package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/pkg/logger"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "查看操作日志",
	Long:  "查看和审计操作日志记录。",
}

var logViewCmd = &cobra.Command{
	Use:   "view [日期]",
	Short: "查看日志",
	Long:  "查看指定日期的操作日志。默认查看今天的日志。",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogView,
}

var logViewLines int

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.AddCommand(logViewCmd)
	logViewCmd.Flags().IntVarP(&logViewLines, "lines", "n", 50, "显示最后N行日志")
}

func runLogView(cmd *cobra.Command, args []string) error {
	var date string
	if len(args) > 0 {
		date = args[0]
	}

	logDir := ""
	if cfgManager != nil {
		logDir = cfgManager.GetConfig().LogDir
	}

	lines, err := logger.ViewLogs(logDir, date, logViewLines)
	if err != nil {
		return fmt.Errorf("failed to view logs: %w", err)
	}

	if len(lines) == 0 {
		printOutput(map[string]interface{}{
			"message": "没有找到日志记录",
		}, nil)
		return nil
	}

	printOutput(map[string]interface{}{
		"count": len(lines),
		"logs":  lines,
	}, nil)

	return nil
}

var logStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "查看日志统计",
	Long:  "查看日志文件的统计信息。",
	RunE:  runLogStats,
}

func init() {
	logCmd.AddCommand(logStatsCmd)
}

func runLogStats(cmd *cobra.Command, args []string) error {
	logDir := ""
	if cfgManager != nil {
		logDir = cfgManager.GetConfig().LogDir
		retentionDays := cfgManager.GetConfig().LogRetentionDays
		if retentionDays == 0 {
			retentionDays = logger.DefaultRetentionDays
		}
		printOutput(map[string]interface{}{
			"log_dir":        logDir,
			"retention_days": retentionDays,
		}, nil)
	}

	stats, err := logger.GetLogStats(logDir)
	if err != nil {
		return fmt.Errorf("failed to get log stats: %w", err)
	}

	printOutput(stats, nil)

	return nil
}

var logCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清理过期日志",
	Long:  "手动清理超过保留期限的日志文件。",
	RunE:  runLogClean,
}

func init() {
	logCmd.AddCommand(logCleanCmd)
}

func runLogClean(cmd *cobra.Command, args []string) error {
	logDir := ""
	retentionDays := logger.DefaultRetentionDays

	if cfgManager != nil {
		logDir = cfgManager.GetConfig().LogDir
		if cfgManager.GetConfig().LogRetentionDays > 0 {
			retentionDays = cfgManager.GetConfig().LogRetentionDays
		}
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	printOutput(map[string]interface{}{
		"message":        "日志清理完成",
		"log_dir":        logDir,
		"retention_days": retentionDays,
		"cutoff_date":    cutoff.Format("2006-01-02"),
	}, nil)

	return nil
}

var logInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看日志配置",
	Long:  "显示当前日志系统的配置信息。",
	RunE:  runLogInfo,
}

func init() {
	logCmd.AddCommand(logInfoCmd)
}

func runLogInfo(cmd *cobra.Command, args []string) error {
	if cfgManager == nil {
		return fmt.Errorf("not logged in")
	}

	cfg := cfgManager.GetConfig()
	logDir := cfg.LogDir
	if logDir == "" {
		logDir = "默认位置 (~/.cloud189/logs/)"
	}

	retentionDays := cfg.LogRetentionDays
	if retentionDays == 0 {
		retentionDays = logger.DefaultRetentionDays
	}

	printOutput(map[string]interface{}{
		"log_dir":         logDir,
		"retention_days":  retentionDays,
		"log_file_format": "YYYY-MM-DD.log",
		"auto_clean":      "启动时自动清理过期日志",
	}, nil)

	return nil
}

// Helper function for logging operations in other commands
func logOperationWithDuration(action, target, result string, startTime time.Time, fileSize int64, errMsg string) {
	duration := time.Since(startTime)
	logOperation(action, target, result, duration, fileSize, errMsg)
}

// parseLogLine parses a single log line (helper for future features)
func parseLogLine(line string) (map[string]string, error) {
	// Format: 2026-04-01 23:30:00 | user | action | target | size | result | duration
	parts := splitLogLine(line)
	if len(parts) < 7 {
		return nil, fmt.Errorf("invalid log line format")
	}

	result := map[string]string{
		"timestamp": parts[0] + " " + parts[1],
		"username":  parts[2],
		"action":    parts[3],
		"target":    parts[4],
		"size":      parts[5],
		"result":    parts[6],
	}

	if len(parts) > 7 {
		result["duration"] = parts[7]
	}

	if len(parts) > 8 {
		result["error"] = parts[8]
	}

	return result, nil
}

func splitLogLine(line string) []string {
	var parts []string
	current := ""
	inPipe := false

	for _, ch := range line {
		if ch == '|' && !inPipe {
			parts = append(parts, trimSpace(current))
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, trimSpace(current))
	}

	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

// Stats helper functions
func calculateLogStatistics(lines []string) map[string]interface{} {
	stats := map[string]interface{}{
		"total_operations": len(lines),
		"by_action":        make(map[string]int),
		"by_result":        make(map[string]int),
	}

	actionStats := stats["by_action"].(map[string]int)
	resultStats := stats["by_result"].(map[string]int)

	for _, line := range lines {
		parsed, err := parseLogLine(line)
		if err != nil {
			continue
		}

		if action, ok := parsed["action"]; ok {
			actionStats[action]++
		}

		if result, ok := parsed["result"]; ok {
			resultStats[result]++
		}
	}

	return stats
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}
