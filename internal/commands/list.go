package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/api"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

var listCmd = &cobra.Command{
	Use:     "ls [路径]",
	Aliases: []string{"list"},
	Short:   "列出文件",
	Long:    "列出指定目录下的文件和文件夹。",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runList,
}

var (
	listLong      bool
	listRecursive bool
	listOrderBy   string
	listDesc      bool
	listPage      int
	listPageSize  int
	listFamily    bool
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listLong, "long", "l", false, "显示详细信息")
	listCmd.Flags().BoolVarP(&listRecursive, "recursive", "r", false, "递归列出")
	listCmd.Flags().StringVar(&listOrderBy, "order-by", "filename", "排序字段 (filename|filesize|lastOpTime)")
	listCmd.Flags().BoolVar(&listDesc, "desc", false, "降序排列")
	listCmd.Flags().IntVarP(&listPage, "page", "n", 1, "页码")
	listCmd.Flags().IntVar(&listPageSize, "page-size", 100, "每页数量")
	listCmd.Flags().BoolVar(&listFamily, "family", false, "家庭云")
}

func runList(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	path := "/"
	if len(args) > 0 {
		path = args[0]
	}

	client := newClient()

	resolver := api.NewPathResolver(client, listFamily)

	folderId, err := resolver.ResolvePath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	files, err := client.ListFiles(folderId, listPage, listPageSize, listOrderBy,
		map[bool]string{true: "desc", false: "asc"}[listDesc], listFamily)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	printOutput(map[string]interface{}{
		"path":      path,
		"files":     files,
		"count":     len(files),
		"page":      listPage,
		"page_size": listPageSize,
	}, nil)

	return nil
}

func parseOrderBy(orderBy string) string {
	switch orderBy {
	case "name", "filename":
		return "filename"
	case "size", "filesize":
		return "filesize"
	case "time", "lastOpTime":
		return "lastOpTime"
	default:
		return "filename"
	}
}

func parseDesc(desc bool) string {
	if desc {
		return "desc"
	}
	return "asc"
}

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <路径>",
	Short: "创建文件夹",
	Long:  "在指定路径创建新文件夹。",
	Args:  cobra.ExactArgs(1),
	RunE:  runMkdir,
}

var mkdirFamily bool

func init() {
	rootCmd.AddCommand(mkdirCmd)
	mkdirCmd.Flags().BoolVar(&mkdirFamily, "family", false, "家庭云")
}

func runMkdir(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	path := args[0]

	if path == "/" || path == "" {
		return fmt.Errorf("invalid folder name")
	}

	client := newClient()
	resolver := api.NewPathResolver(client, mkdirFamily)

	parentPath, folderName := pathResolverGetParent(path)

	parentId, err := resolver.ResolvePathWithCreate(parentPath, false)
	if err != nil {
		logOperation("mkdir", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to resolve parent path: %w", err)
	}

	folder, err := client.CreateFolder(parentId, folderName, mkdirFamily)
	if err != nil {
		logOperation("mkdir", path, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to create folder: %w", err)
	}

	logOperation("mkdir", path, "success", time.Since(startTime), 0, "")

	printOutput(map[string]interface{}{
		"id":      folder.ID,
		"name":    folder.Name,
		"message": "文件夹创建成功",
	}, nil)

	return nil
}

func pathResolverGetParent(pathStr string) (string, string) {
	pathStr = strings.TrimSuffix(pathStr, "/")

	lastSlash := strings.LastIndex(pathStr, "/")

	if lastSlash == -1 {
		return "/", pathStr
	}

	if lastSlash == 0 {
		return "/", pathStr[1:]
	}

	return pathStr[:lastSlash], pathStr[lastSlash+1:]
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看容量信息",
	Long:  "查看个人云和家庭云的容量使用情况。",
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	client := newClient()

	capacity, err := client.GetCapacityInfo()
	if err != nil {
		return fmt.Errorf("failed to get capacity info: %w", err)
	}

	if capacity.ResCode != 0 {
		return fmt.Errorf("API error: %s (code: %d)", capacity.ResMessage, capacity.ResCode)
	}

	printOutput(map[string]interface{}{
		"account": capacity.Account,
		"personal": map[string]interface{}{
			"total":    capacity.CloudCapacityInfo.TotalSize,
			"used":     capacity.CloudCapacityInfo.UsedSize,
			"free":     capacity.CloudCapacityInfo.FreeSize,
			"total_gb": float64(capacity.CloudCapacityInfo.TotalSize) / 1024 / 1024 / 1024,
			"used_gb":  float64(capacity.CloudCapacityInfo.UsedSize) / 1024 / 1024 / 1024,
		},
		"family": map[string]interface{}{
			"total":    capacity.FamilyCapacityInfo.TotalSize,
			"used":     capacity.FamilyCapacityInfo.UsedSize,
			"free":     capacity.FamilyCapacityInfo.FreeSize,
			"total_gb": float64(capacity.FamilyCapacityInfo.TotalSize) / 1024 / 1024 / 1024,
			"used_gb":  float64(capacity.FamilyCapacityInfo.UsedSize) / 1024 / 1024 / 1024,
		},
	}, nil)

	return nil
}

var familyCmd = &cobra.Command{
	Use:   "family",
	Short: "家庭云操作",
	Long:  "查看和切换家庭云。",
}

var familyListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出家庭云",
	RunE:  runFamilyList,
}

func init() {
	rootCmd.AddCommand(familyCmd)
	familyCmd.AddCommand(familyListCmd)
}

func runFamilyList(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	client := newClient()

	families, err := client.GetFamilyList()
	if err != nil {
		return fmt.Errorf("failed to get family list: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(families))
	for _, f := range families {
		result = append(result, map[string]interface{}{
			"family_id":   strconv.FormatInt(f.FamilyID, 10),
			"remark_name": f.RemarkName,
			"create_time": f.CreateTime,
			"user_role":   f.UserRole,
		})
	}

	printOutput(map[string]interface{}{
		"count":    len(families),
		"families": result,
	}, nil)

	return nil
}

var familyUseCmd = &cobra.Command{
	Use:   "use <家庭云ID>",
	Short: "切换家庭云",
	Long:  "切换到指定的家庭云，后续操作将使用该家庭云。",
	Args:  cobra.ExactArgs(1),
	RunE:  runFamilyUse,
}

func init() {
	familyCmd.AddCommand(familyUseCmd)
}

func runFamilyUse(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	familyID := args[0]

	// 验证家庭云ID是否有效
	client := newClient()
	families, err := client.GetFamilyList()
	if err != nil {
		return fmt.Errorf("failed to get family list: %w", err)
	}

	found := false
	var familyName string
	for _, f := range families {
		if strconv.FormatInt(f.FamilyID, 10) == familyID {
			found = true
			familyName = f.RemarkName
			break
		}
	}

	if !found {
		return fmt.Errorf("family ID '%s' not found", familyID)
	}

	// 保存到配置
	if err := cfgManager.SetFamilyID(familyID); err != nil {
		return fmt.Errorf("failed to set family ID: %w", err)
	}

	logOperation("family-use", familyID, "success", 0, 0, "")

	printOutput(map[string]interface{}{
		"family_id":   familyID,
		"family_name": familyName,
		"message":     "家庭云切换成功",
	}, nil)

	return nil
}

var familySaveCmd = &cobra.Command{
	Use:   "save <家庭云文件路径> <个人云目标路径>",
	Short: "保存家庭云文件到个人云",
	Long:  "将家庭云的文件或文件夹转存到个人云指定路径。",
	Args:  cobra.ExactArgs(2),
	RunE:  runFamilySave,
}

func init() {
	familyCmd.AddCommand(familySaveCmd)
}

func runFamilySave(cmd *cobra.Command, args []string) error {
	if !cfgManager.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	srcPath := args[0]
	dstPath := args[1]

	// 获取家庭云ID
	if cfgManager.GetConfig().FamilyID == "" {
		return fmt.Errorf("no family cloud selected, please use 'family use <id>' first")
	}

	client := newClient()

	// 解析家庭云源路径
	srcResolver := api.NewPathResolver(client, true)
	srcParentPath, srcFileName := pathResolverGetParent(srcPath)
	srcParentId, err := srcResolver.ResolvePath(srcParentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	srcFiles, err := client.ListFiles(srcParentId, 1, 1000, "filename", "asc", true)
	if err != nil {
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
		return fmt.Errorf("source file '%s' not found in family cloud", srcFileName)
	}

	// 解析个人云目标路径
	dstResolver := api.NewPathResolver(client, false)
	dstParentId, err := dstResolver.ResolvePath(dstPath)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// 创建目标文件夹对象
	dstDir := &types.File{
		ID:    dstParentId,
		IsDir: true,
	}

	// 执行转存
	err = client.SaveFamilyFileToPersonCloud(cfgManager.GetConfig().FamilyID, srcFile, dstDir, false)
	if err != nil {
		return fmt.Errorf("failed to save family file to personal cloud: %w", err)
	}

	logOperation("family-save", srcPath+" -> "+dstPath, "success", 0, srcFile.Size, "")

	printOutput(map[string]interface{}{
		"source_path":      srcPath,
		"destination_path": dstPath,
		"file_name":        srcFile.Name,
		"file_size":        srcFile.Size,
		"message":          "转存成功",
	}, nil)

	return nil
}
