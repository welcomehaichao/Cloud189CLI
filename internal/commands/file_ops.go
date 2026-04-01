package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

var renameCmd = &cobra.Command{
	Use:   "rename <文件路径> <新名称>",
	Short: "重命名文件或文件夹",
	Long:  "重命名指定路径的文件或文件夹。",
	Args:  cobra.ExactArgs(2),
	RunE:  runRename,
}

var renameFamily bool

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.Flags().BoolVar(&renameFamily, "family", false, "家庭云")
}

func runRename(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	path := args[0]
	newName := args[1]

	client := newClient()
	resolver := api.NewPathResolver(client, renameFamily)

	parentPath, oldName := pathResolverGetParent(path)

	parentId, err := resolver.ResolvePath(parentPath)
	if err != nil {
		logOperation("rename", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	files, err := client.ListFiles(parentId, 1, 1000, "filename", "asc", renameFamily)
	if err != nil {
		logOperation("rename", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list files: %w", err)
	}

	var targetFile *types.File
	for i := range files {
		if files[i].Name == oldName {
			targetFile = &files[i]
			break
		}
	}

	if targetFile == nil {
		logOperation("rename", path, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", oldName))
		return fmt.Errorf("file or folder '%s' not found", oldName)
	}

	if targetFile.IsDir {
		err = client.RenameFolder(targetFile.ID, newName, renameFamily)
	} else {
		err = client.RenameFile(targetFile.ID, newName, renameFamily)
	}

	if err != nil {
		logOperation("rename", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to rename: %w", err)
	}

	logOperation("rename", path, "success", time.Since(startTime), 0, "")

	printOutput(map[string]interface{}{
		"id":       targetFile.ID,
		"old_name": oldName,
		"new_name": newName,
		"message":  "重命名成功",
	}, nil)

	return nil
}

var rmCmd = &cobra.Command{
	Use:     "rm <文件路径>",
	Aliases: []string{"delete", "remove"},
	Short:   "删除文件或文件夹",
	Long:    "删除指定路径的文件或文件夹，删除后可在回收站找回。",
	Args:    cobra.ExactArgs(1),
	RunE:    runRm,
}

var rmForce bool
var rmPermanent bool
var rmFamily bool

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "强制删除，不确认")
	rmCmd.Flags().BoolVar(&rmPermanent, "permanent", false, "永久删除（清空回收站）")
	rmCmd.Flags().BoolVar(&rmFamily, "family", false, "家庭云")
}

func runRm(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	path := args[0]

	if !rmForce {
		fmt.Printf("确定要删除 '%s' 吗？(y/n): ", path)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			logOperation("delete", path, "cancelled", time.Since(startTime), 0, "")
			return fmt.Errorf("operation cancelled")
		}
	}

	client := newClient()
	resolver := api.NewPathResolver(client, rmFamily)

	parentPath, fileName := pathResolverGetParent(path)

	parentId, err := resolver.ResolvePath(parentPath)
	if err != nil {
		logOperation("delete", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	files, err := client.ListFiles(parentId, 1, 1000, "filename", "asc", rmFamily)
	if err != nil {
		logOperation("delete", path, "failed", time.Since(startTime), 0, err.Error())
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
		logOperation("delete", path, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", fileName))
		return fmt.Errorf("file or folder '%s' not found", fileName)
	}

	err = client.Delete(targetFile, rmFamily)
	if err != nil {
		logOperation("delete", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to delete: %w", err)
	}

	logOperation("delete", path, "success", time.Since(startTime), 0, "")

	printOutput(map[string]interface{}{
		"id":      targetFile.ID,
		"name":    fileName,
		"message": "删除成功",
	}, nil)

	return nil
}

var mvCmd = &cobra.Command{
	Use:     "mv <源路径> <目标路径>",
	Aliases: []string{"move"},
	Short:   "移动文件或文件夹",
	Long:    "将文件或文件夹移动到指定位置。",
	Args:    cobra.ExactArgs(2),
	RunE:    runMv,
}

var mvFamily bool

func init() {
	rootCmd.AddCommand(mvCmd)
	mvCmd.Flags().BoolVar(&mvFamily, "family", false, "家庭云")
}

func runMv(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	srcPath := args[0]
	dstPath := args[1]

	client := newClient()
	resolver := api.NewPathResolver(client, mvFamily)

	srcParentPath, srcFileName := pathResolverGetParent(srcPath)

	srcParentId, err := resolver.ResolvePath(srcParentPath)
	if err != nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve source parent path: %w", err)
	}

	srcFiles, err := client.ListFiles(srcParentId, 1, 1000, "filename", "asc", mvFamily)
	if err != nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list source files: %w", err)
	}

	var srcFile *types.File
	for i := range srcFiles {
		if srcFiles[i].Name == srcFileName {
			srcFile = &srcFiles[i]
			break
		}
	}

	if srcFile == nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", srcFileName))
		return fmt.Errorf("source file or folder '%s' not found", srcFileName)
	}

	dstParentPath, dstFileName := pathResolverGetParent(dstPath)
	if dstFileName == "" || dstFileName == srcFileName {
		dstFileName = srcFileName
	}

	dstParentId, err := resolver.ResolvePath(dstParentPath)
	if err != nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve destination parent path: %w", err)
	}

	dstFiles, err := client.ListFiles(dstParentId, 1, 1000, "filename", "asc", mvFamily)
	if err != nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list destination files: %w", err)
	}

	var dstDir *types.File
	dstDir = &types.File{
		ID:    dstParentId,
		Name:  dstFileName,
		IsDir: true,
	}

	for i := range dstFiles {
		if dstFiles[i].Name == dstFileName {
			dstDir = &dstFiles[i]
			break
		}
	}

	err = client.Move(srcFile, dstDir, mvFamily)
	if err != nil {
		logOperation("move", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to move: %w", err)
	}

	logOperation("move", srcPath+" -> "+dstPath, "success", time.Since(startTime), 0, "")

	printOutput(map[string]interface{}{
		"id":      srcFile.ID,
		"message": "移动成功",
		"name":    srcFileName,
	}, nil)

	return nil
}

var cpCmd = &cobra.Command{
	Use:     "cp <源路径> <目标路径>",
	Aliases: []string{"copy"},
	Short:   "复制文件或文件夹",
	Long:    "将文件或文件夹复制到指定位置。",
	Args:    cobra.ExactArgs(2),
	RunE:    runCp,
}

var cpFamily bool

func init() {
	rootCmd.AddCommand(cpCmd)
	cpCmd.Flags().BoolVar(&cpFamily, "family", false, "家庭云")
}

func runCp(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	srcPath := args[0]
	dstPath := args[1]

	client := newClient()
	resolver := api.NewPathResolver(client, cpFamily)

	srcParentPath, srcFileName := pathResolverGetParent(srcPath)

	srcParentId, err := resolver.ResolvePath(srcParentPath)
	if err != nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve source parent path: %w", err)
	}

	srcFiles, err := client.ListFiles(srcParentId, 1, 1000, "filename", "asc", cpFamily)
	if err != nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list source files: %w", err)
	}

	var srcFile *types.File
	for i := range srcFiles {
		if srcFiles[i].Name == srcFileName {
			srcFile = &srcFiles[i]
			break
		}
	}

	if srcFile == nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", srcFileName))
		return fmt.Errorf("source file or folder '%s' not found", srcFileName)
	}

	dstParentPath, dstFileName := pathResolverGetParent(dstPath)
	if dstFileName == "" || dstFileName == srcFileName {
		dstFileName = srcFileName
	}

	dstParentId, err := resolver.ResolvePath(dstParentPath)
	if err != nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve destination parent path: %w", err)
	}

	dstFiles, err := client.ListFiles(dstParentId, 1, 1000, "filename", "asc", cpFamily)
	if err != nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to list destination files: %w", err)
	}

	var dstDir *types.File
	dstDir = &types.File{
		ID:    dstParentId,
		Name:  dstFileName,
		IsDir: true,
	}

	for i := range dstFiles {
		if dstFiles[i].Name == dstFileName {
			dstDir = &dstFiles[i]
			break
		}
	}

	err = client.Copy(srcFile, dstDir, cpFamily)
	if err != nil {
		logOperation("copy", srcPath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to copy: %w", err)
	}

	logOperation("copy", srcPath+" -> "+dstPath, "success", time.Since(startTime), 0, "")

	printOutput(map[string]interface{}{
		"id":      srcFile.ID,
		"message": "复制成功",
		"name":    srcFileName,
	}, nil)

	return nil
}

var getUrlCmd = &cobra.Command{
	Use:   "get-url <文件路径>",
	Short: "获取文件下载链接",
	Long:  "获取指定文件的下载链接及详细信息，包括过期时间。",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetUrl,
}

var getUrlFamily bool
var getUrlRaw bool

func init() {
	rootCmd.AddCommand(getUrlCmd)
	getUrlCmd.Flags().BoolVar(&getUrlFamily, "family", false, "家庭云")
	getUrlCmd.Flags().BoolVar(&getUrlRaw, "raw", false, "仅输出下载链接（用于脚本）")
}

func runGetUrl(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	filePath := args[0]

	client := newClient()
	resolver := api.NewPathResolver(client, getUrlFamily)

	parentPath, fileName := pathResolverGetParent(filePath)

	parentId, err := resolver.ResolvePath(parentPath)
	if err != nil {
		logOperation("get-url", filePath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	files, err := client.ListFiles(parentId, 1, 1000, "filename", "asc", getUrlFamily)
	if err != nil {
		logOperation("get-url", filePath, "failed", time.Since(startTime), 0, err.Error())
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
		logOperation("get-url", filePath, "failed", time.Since(startTime), 0, fmt.Sprintf("file '%s' not found", fileName))
		return fmt.Errorf("file '%s' not found", fileName)
	}

	if targetFile.IsDir {
		logOperation("get-url", filePath, "failed", time.Since(startTime), 0, "cannot get download URL for a folder")
		return fmt.Errorf("cannot get download URL for a folder, please specify a file")
	}

	urlInfo, err := client.GetDownloadURLInfo(targetFile.ID, getUrlFamily)
	if err != nil {
		logOperation("get-url", filePath, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	logOperation("get-url", filePath, "success", time.Since(startTime), targetFile.Size, "")

	if getUrlRaw {
		fmt.Println(urlInfo.DownloadURL)
	} else {
		expireTimeStr := ""
		if !urlInfo.ExpireTime.IsZero() {
			expireTimeStr = urlInfo.ExpireTime.Format("2006-01-02 15:04:05")
		}

		printOutput(map[string]interface{}{
			"file_id":      urlInfo.FileID,
			"file_name":    targetFile.Name,
			"file_size":    targetFile.Size,
			"md5":          targetFile.MD5,
			"download_url": urlInfo.DownloadURL,
			"expire_time":  expireTimeStr,
			"expired":      urlInfo.Expired,
		}, nil)
	}

	return nil
}
