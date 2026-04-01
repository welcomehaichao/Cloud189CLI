package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录天翼云盘",
	Long:  "使用用户名密码或二维码登录天翼云盘。",
	RunE:  runLogin,
}

var (
	username string
	password string
	qrLogin  bool
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "用户名")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "密码")
	loginCmd.Flags().BoolVarP(&qrLogin, "qr", "q", false, "二维码登录")
}

func runLogin(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	client := newClient()

	var session *types.Session
	var err error

	if qrLogin {
		session, err = client.LoginByQRCode()
	} else {
		if username == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("请输入用户名: ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)
		}

		if password == "" {
			fmt.Print("请输入密码: ")
			passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			password = string(passwordBytes)
		}

		session, err = client.Login(username, password)
	}

	duration := time.Since(startTime)

	if err != nil {
		logOperation("login", "-", "failed", duration, 0, err.Error())
		return fmt.Errorf("login failed: %w", err)
	}

	if err := cfgManager.SetSession(session); err != nil {
		logOperation("login", session.LoginName, "failed", duration, 0, err.Error())
		return fmt.Errorf("failed to save session: %w", err)
	}

	logOperation("login", session.LoginName, "success", duration, 0, "")

	printOutput(map[string]string{
		"message":  "登录成功",
		"username": session.LoginName,
	}, nil)

	return nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	username := cfgManager.GetConfig().Username

	if err := cfgManager.Clear(); err != nil {
		logOperation("logout", username, "failed", time.Since(startTime), 0, err.Error())
		return fmt.Errorf("failed to clear config: %w", err)
	}

	logOperation("logout", username, "success", time.Since(startTime), 0, "")

	printOutput(map[string]string{
		"message": "已退出登录",
	}, nil)

	return nil
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "查看当前登录用户",
	RunE:  runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	cfg := cfgManager.GetConfig()

	if !cfgManager.IsLoggedIn() {
		printOutput(nil, fmt.Errorf("not logged in"))
		return nil
	}

	printOutput(map[string]string{
		"username": cfg.Username,
	}, nil)

	return nil
}
