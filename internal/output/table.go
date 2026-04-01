package output

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func PrintTable(out *Output) error {
	if out.Success && out.Data != nil {
		return printDataTable(out.Data)
	}

	if out.Error != nil {
		fmt.Printf("错误: [%s] %s\n", out.Error.Code, out.Error.Message)
		if len(out.Error.Details) > 0 {
			fmt.Println("详细信息:")
			for k, v := range out.Error.Details {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}

	return nil
}

func printDataTable(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Map {
		return printMapTable(data)
	}

	fmt.Println(formatValue(data))
	return nil
}

func printMapTable(data interface{}) error {
	m, ok := data.(map[string]interface{})
	if !ok {
		fmt.Println(formatValue(data))
		return nil
	}

	if account, exists := m["account"]; exists {
		fmt.Printf("\n账号: %v\n\n", account)
	}

	if personal, exists := m["personal"]; exists {
		fmt.Println("=== 个人云 ===")
		printCapacityTable(personal)
		fmt.Println()
	}

	if family, exists := m["family"]; exists {
		fmt.Println("=== 家庭云 ===")
		printCapacityTable(family)
		fmt.Println()
	}

	if files, exists := m["files"]; exists {
		fmt.Println("文件列表:")
		printFilesTable(files)
	}

	return nil
}

func printCapacityTable(data interface{}) {
	m, ok := data.(map[string]interface{})
	if !ok {
		fmt.Println(formatValue(data))
		return
	}

	total := getInt64(m, "total")
	used := getInt64(m, "used")
	free := getInt64(m, "free")
	totalGB := getFloat64(m, "total_gb")
	usedGB := getFloat64(m, "used_gb")

	fmt.Printf("%-15s %20s %15s\n", "类型", "大小", "GB")
	fmt.Println(strings.Repeat("-", 55))
	fmt.Printf("%-15s %20d %15.2f\n", "总容量", total, totalGB)
	fmt.Printf("%-15s %20d %15.2f\n", "已使用", used, usedGB)
	fmt.Printf("%-15s %20d %15.2f\n", "剩余空间", free, float64(free)/1024/1024/1024)
	fmt.Printf("%-15s %20s %14.1f%%\n", "使用率", "", float64(used)/float64(total)*100)
}

func printFilesTable(data interface{}) {
	slice, ok := data.([]interface{})
	if !ok {
		fmt.Println(formatValue(data))
		return
	}

	if len(slice) == 0 {
		fmt.Println("(空)")
		return
	}

	fmt.Printf("\n%-20s %-10s %-15s %-20s\n", "名称", "大小", "类型", "修改时间")
	fmt.Println(strings.Repeat("-", 70))

	for _, item := range slice {
		if m, ok := item.(map[string]interface{}); ok {
			name := getString(m, "name")
			if len(name) > 18 {
				name = name[:15] + "..."
			}

			size := getInt64(m, "size")
			isDir := getBool(m, "is_dir")
			modified := getString(m, "modified")

			fileType := "文件"
			sizeStr := formatSize(size)
			if isDir {
				fileType = "文件夹"
				sizeStr = "-"
			}

			fmt.Printf("%-20s %-10s %-15s %-20s\n", name, sizeStr, fileType, modified)
		}
	}
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, exists := m[key]; exists {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, exists := m[key]; exists {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, exists := m[key]; exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, exists := m[key]; exists {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func printSimpleTable(data interface{}) {
	fmt.Fprint(os.Stdout, formatValue(data))
}
