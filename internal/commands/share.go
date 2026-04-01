package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

var shareCmd = &cobra.Command{
	Use:   "share <文件路径>",
	Short: "创建文件分享链接",
	Long:  "为指定的文件或文件夹创建分享链接。",
	Args:  cobra.ExactArgs(1),
	RunE:  runShare,
}

var shareFamily bool
var shareExpireDays int
var shareAccessCode string

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.Flags().BoolVar(&shareFamily, "family", false, "家庭云")
	shareCmd.Flags().IntVarP(&shareExpireDays, "expire", "e", 0, "分享有效期（天），0表示永久")
	shareCmd.Flags().StringVarP(&shareAccessCode, "code", "c", "", "提取码（留空则自动生成）")
}

func runShare(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	path := args[0]

	client := newClient()
	resolver := api.NewPathResolver(client, shareFamily)

	parentPath, fileName := pathResolverGetParent(path)

	parentId, err := resolver.ResolvePath(parentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	files, err := client.ListFiles(parentId, 1, 1000, "filename", "asc", shareFamily)
	if err != nil {
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
		return fmt.Errorf("file or folder '%s' not found", fileName)
	}

	shareInfo, err := client.CreateShareLink(targetFile.ID, targetFile.IsDir, shareExpireDays, shareAccessCode, shareFamily)
	if err != nil {
		return fmt.Errorf("failed to create share link: %w", err)
	}

	logOperation("share", path, "success", 0, 0, "")

	printOutput(map[string]interface{}{
		"share_id":    shareInfo.ShareId,
		"share_link":  shareInfo.ShareLink,
		"access_code": shareInfo.AccessCode,
		"file_name":   targetFile.Name,
		"file_id":     targetFile.ID,
		"is_folder":   targetFile.IsDir,
		"expire_days": shareExpireDays,
		"message":     "分享链接创建成功",
	}, nil)

	return nil
}

var shareCancelCmd = &cobra.Command{
	Use:   "share-cancel <分享ID>",
	Short: "取消分享",
	Long:  "取消指定的分享链接。",
	Args:  cobra.ExactArgs(1),
	RunE:  runShareCancel,
}

func init() {
	rootCmd.AddCommand(shareCancelCmd)
}

func runShareCancel(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	shareId := args[0]

	client := newClient()

	err := client.CancelShare(shareId)
	if err != nil {
		return fmt.Errorf("failed to cancel share: %w", err)
	}

	logOperation("share-cancel", shareId, "success", 0, 0, "")

	printOutput(map[string]interface{}{
		"share_id": shareId,
		"message":  "分享已取消",
	}, nil)

	return nil
}
