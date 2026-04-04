package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

var downloadCmd = &cobra.Command{
	Use:   "download <云端文件路径> <本地路径>",
	Short: "从天翼云盘下载文件",
	Long:  "从天翼云盘下载文件到本地，支持断点续传。",
	Args:  cobra.ExactArgs(2),
	RunE:  runDownload,
}

var (
	downloadFamily   bool
	downloadResume   bool
	downloadProgress bool
)

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolVar(&downloadFamily, "family", false, "从家庭云下载")
	downloadCmd.Flags().BoolVar(&downloadResume, "resume", false, "断点续传")
	downloadCmd.Flags().BoolVar(&downloadProgress, "progress", true, "显示下载进度")
}

func runDownload(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	cloudPath := args[0]
	localPath := args[1]

	client := newClient()
	resolver := api.NewPathResolver(client, downloadFamily)

	parentPath, fileName := pathResolverGetParent(cloudPath)

	parentId, err := resolver.ResolvePath(parentPath)
	if err != nil {
		logOperation("download", cloudPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	files, err := client.ListFiles(parentId, 1, 1000, "filename", "asc", downloadFamily)
	if err != nil {
		logOperation("download", cloudPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list files: %w", err)
	}

	var targetFile *types.File
	for i := range files {
		if files[i].Name == fileName {
			targetFile = &files[i]
			break
		}
	}

	if targetFile == nil {
		logOperation("download", cloudPath, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", fileName))
		return fmt.Errorf("file '%s' not found", fileName)
	}

	if targetFile.IsDir {
		logOperation("download", cloudPath, "failed", time.Since(startTime), 0, "cannot download a folder")
		return fmt.Errorf("cannot download a folder, please specify a file")
	}

	localFileInfo, err := os.Stat(localPath)
	if err == nil && localFileInfo.IsDir() {
		localPath = filepath.Join(localPath, fileName)
	}

	var progressCallback func(percent float64)
	if downloadProgress {
		progressCallback = func(percent float64) {
			fmt.Printf("\r下载进度: %.2f%%", percent)
			if percent >= 100.0 {
				fmt.Println()
			}
		}
	}

	ctx := context.Background()
	var downloadErr error

	if downloadResume {
		downloadErr = client.DownloadFileWithResume(ctx, targetFile.ID, localPath, progressCallback, downloadFamily)
	} else {
		downloadErr = client.DownloadFile(ctx, targetFile.ID, localPath, progressCallback, downloadFamily)
	}

	if downloadErr != nil {
		logOperation("download", cloudPath+" -> "+localPath, "failed", time.Since(startTime), targetFile.Size, downloadErr.Error())
		return fmt.Errorf("failed to download file: %w", downloadErr)
	}

	logOperation("download", cloudPath+" -> "+localPath, "success", time.Since(startTime), targetFile.Size, "")

	if downloadProgress {
		fmt.Printf("下载成功: %s -> %s\n", fileName, localPath)
	}

	printOutput(map[string]interface{}{
		"file_id":    targetFile.ID,
		"file_name":  targetFile.Name,
		"file_size":  targetFile.Size,
		"local_path": localPath,
		"md5":        targetFile.MD5,
		"message":    "下载成功",
	}, nil)

	return nil
}
