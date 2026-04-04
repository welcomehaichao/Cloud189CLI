package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <本地文件> <云端路径>",
	Short: "上传文件到天翼云盘",
	Long:  "上传本地文件到天翼云盘指定路径，支持大文件分片上传和断点续传。",
	Args:  cobra.ExactArgs(2),
	RunE:  runUpload,
}

var (
	uploadFamily   bool
	uploadStream   bool
	uploadResume   bool
	uploadProgress bool
)

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().BoolVar(&uploadFamily, "family", false, "上传到家庭云")
	uploadCmd.Flags().BoolVar(&uploadStream, "stream", false, "使用Stream分片上传（推荐用于大文件）")
	uploadCmd.Flags().BoolVar(&uploadResume, "resume", false, "断点续传（仅Stream模式支持）")
	uploadCmd.Flags().BoolVar(&uploadProgress, "progress", true, "显示上传进度")
}

func runUpload(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	localPath := args[0]
	cloudPath := args[1]

	if _, err := os.Stat(localPath); err != nil {
		logOperation("upload", localPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("local file not found: %w", err)
	}

	fileInfo, err := os.Stat(localPath)
	if err != nil {
		logOperation("upload", localPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.IsDir() {
		logOperation("upload", localPath, "failed", time.Since(startTime), 0, "cannot upload directory")
		return fmt.Errorf("cannot upload a directory, please specify a file")
	}

	fileSize := fileInfo.Size()

	client := newClient()

	var progressCallback func(percent float64)
	if uploadProgress {
		progressCallback = func(percent float64) {
			fmt.Printf("\r上传进度: %.2f%%", percent)
			if percent >= 100.0 {
				fmt.Println()
			}
		}
	}

	ctx := context.Background()
	var result *types.File
	var uploadErr error

	if uploadStream || uploadResume {
		if uploadResume && !uploadStream {
			fmt.Println("提示: 断点续传仅在Stream模式下有效，自动启用Stream模式")
			uploadStream = true
		}

		result, uploadErr = client.StreamUploadWithResume(ctx, localPath, cloudPath, progressCallback, uploadFamily, uploadResume)
	} else {
		result, uploadErr = client.UploadFile(ctx, localPath, cloudPath, progressCallback, uploadFamily)
	}

	if uploadErr != nil {
		logOperation("upload", localPath+" -> "+cloudPath, "failed", time.Since(startTime), fileSize, uploadErr.Error())
		return fmt.Errorf("failed to upload file: %w", uploadErr)
	}

	logOperation("upload", localPath+" -> "+cloudPath, "success", time.Since(startTime), fileSize, "")

	if uploadProgress && result != nil {
		fmt.Printf("上传成功: %s (ID: %s, MD5: %s)\n", result.Name, result.ID, result.MD5)
	}

	printOutput(map[string]interface{}{
		"file_id":   result.ID,
		"file_name": result.Name,
		"file_size": result.Size,
		"md5":       result.MD5,
		"message":   "上传成功",
	}, nil)

	return nil
}
