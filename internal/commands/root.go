package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/internal/config"
	"github.com/yuhaichao/cloud189-cli/internal/output"
	"github.com/yuhaichao/cloud189-cli/pkg/logger"
)

var (
	cfgManager   *config.Manager
	outputFormat output.OutputFormat
	auditLog     *logger.AuditLogger
	version      string
	buildTime    string
)

func SetVersionInfo(v, bt string) {
	version = v
	buildTime = bt
}

var rootCmd = &cobra.Command{
	Use:   "cloud189",
	Short: "天翼云盘命令行工具",
	Long:  "cloud189是一个功能完整的天翼云盘命令行工具，支持文件上传、下载、管理等功能。",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfgManager, err = config.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		if err := cfgManager.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// 初始化审计日志
		if cfgManager.IsLoggedIn() {
			auditLog, err = logger.NewAuditLogger(
				cfgManager.GetConfig().LogDir,
				cfgManager.GetConfig().LogRetentionDays,
				cfgManager.GetConfig().Username,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if auditLog != nil {
			auditLog.Close()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP((*string)(&outputFormat), "output", "o", "json", "输出格式 (json|yaml|table)")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cloud189 CLI v%s\n", version)
			fmt.Printf("Build Time: %s\n", buildTime)
			fmt.Printf("Go Version: %s\n", "1.23.3")
		},
	}
	rootCmd.AddCommand(versionCmd)
}

func printOutput(data interface{}, err error) {
	var out *output.Output

	if err != nil {
		out = output.NewErrorOutput("ERROR", err.Error())
	} else {
		out = output.NewOutput(true, data)
	}

	switch outputFormat {
	case output.FormatJSON:
		if err := output.PrintJSON(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print output: %v\n", err)
		}
	case output.FormatYAML:
		if err := output.PrintYAML(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print output: %v\n", err)
		}
	case output.FormatTable:
		if err := output.PrintTable(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print output: %v\n", err)
		}
	default:
		if err := output.PrintJSON(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print output: %v\n", err)
		}
	}
}

// newClient 创建支持Session自动刷新的API客户端
func newClient() *api.Client {
	return api.NewClientWithManager(cfgManager)
}

// logOperation 记录操作日志的辅助函数
func logOperation(action, target, result string, duration time.Duration, fileSize int64, errMsg string) {
	if auditLog == nil {
		return
	}

	entry := &logger.LogEntry{
		Timestamp: time.Now(),
		Username:  cfgManager.GetConfig().Username,
		Action:    action,
		Target:    target,
		FileSize:  fileSize,
		Result:    result,
		Duration:  duration,
		Error:     errMsg,
	}

	if err := auditLog.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to log operation: %v\n", err)
	}
}
