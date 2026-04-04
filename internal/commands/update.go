package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/internal/updater"
)

func init() {
	var checkOnly bool
	var force bool

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "更新到最新版本",
		Long:  "检查并更新cloud189 CLI到最新版本",
		Run: func(cmd *cobra.Command, args []string) {
			u := updater.NewUpdater(version)
			u.CheckOnly = checkOnly
			u.Force = force

			result, err := u.CheckForUpdate()
			if err != nil {
				fmt.Fprintf(os.Stderr, "更新失败: %v\n", err)
				os.Exit(1)
			}

			if checkOnly {
				fmt.Println(result.Message)
			} else {
				fmt.Println(result.Message)
				if result.Updated {
					fmt.Printf("当前版本: %s\n", result.LatestVersion)
					fmt.Println("更新完成！")
				}
			}
		},
	}

	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "仅检查是否有新版本")
	updateCmd.Flags().BoolVar(&force, "force", false, "强制更新（即使版本相同）")

	rootCmd.AddCommand(updateCmd)
}
