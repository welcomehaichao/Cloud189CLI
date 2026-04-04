package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/yuhaichao/cloud189-cli/pkg/utils"
)

type Updater struct {
	CurrentVersion string
	Force          bool
	CheckOnly      bool
	TargetVersion  string
}

type UpdateResult struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
	Message        string
}

func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
	}
}

func (u *Updater) CheckForUpdate() (*UpdateResult, error) {
	release, err := GetLatestRelease()
	if err != nil {
		return nil, err
	}

	latestVersion := release.TagName

	isOlderOrEqual, err := CompareVersions(u.CurrentVersion, latestVersion)
	if err != nil {
		return nil, err
	}

	result := &UpdateResult{
		CurrentVersion: u.CurrentVersion,
		LatestVersion:  latestVersion,
		Updated:        false,
	}

	if isOlderOrEqual && !u.Force {
		result.Message = fmt.Sprintf("已是最新版本 %s", u.CurrentVersion)
		return result, nil
	}

	if u.CheckOnly {
		result.Message = fmt.Sprintf("发现新版本: %s (当前: %s)", latestVersion, u.CurrentVersion)
		return result, nil
	}

	return u.performUpdate(release)
}

func (u *Updater) performUpdate(release *GitHubRelease) (*UpdateResult, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	assetName, assetURL, err := FindAsset(release, osName, arch)
	if err != nil {
		return nil, err
	}

	fmt.Printf("正在下载 %s...\n", assetName)

	tmpDir, err := u.downloadAsset(assetName, assetURL)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}

	defer os.RemoveAll(tmpDir)

	currentBinary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取当前二进制路径失败: %w", err)
	}

	fmt.Println("正在备份旧版本...")
	backupPath := currentBinary + ".backup"
	if err := u.backupBinary(currentBinary, backupPath); err != nil {
		return nil, fmt.Errorf("备份失败: %w", err)
	}

	fmt.Println("正在安装新版本...")
	if err := u.installBinary(tmpDir, currentBinary, osName); err != nil {
		fmt.Println("安装失败，正在恢复旧版本...")
		if restoreErr := u.restoreBackup(backupPath, currentBinary); restoreErr != nil {
			fmt.Printf("恢复失败: %v\n", restoreErr)
		}
		return nil, fmt.Errorf("安装失败: %w", err)
	}

	fmt.Println("清理备份文件...")
	os.Remove(backupPath)

	return &UpdateResult{
		CurrentVersion: u.CurrentVersion,
		LatestVersion:  release.TagName,
		Updated:        true,
		Message:        fmt.Sprintf("已成功更新至 %s", release.TagName),
	}, nil
}

func (u *Updater) downloadAsset(assetName, assetURL string) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "cloud189-update-"+utils.GenerateUUID())
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	assetPath := filepath.Join(tmpDir, assetName)

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(assetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(assetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	size := resp.ContentLength
	downloaded := int64(0)

	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)

			if size > 0 {
				percent := int(float64(downloaded) / float64(size) * 100)
				fmt.Printf("\r下载进度: %d%% (%d/%d bytes)", percent, downloaded, size)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	fmt.Println("\n下载完成")
	return tmpDir, nil
}

func (u *Updater) backupBinary(src, dst string) error {
	return copyFile(src, dst)
}

func (u *Updater) restoreBackup(src, dst string) error {
	return copyFile(src, dst)
}

func (u *Updater) installBinary(tmpDir, targetBinary, osName string) error {
	assetName := ""
	for _, file := range []string{"cloud189-linux-amd64.tar.gz", "cloud189-darwin-amd64.tar.gz", "cloud189-darwin-arm64.tar.gz", "cloud189-windows-amd64.zip"} {
		if strings.Contains(file, osName) {
			assetName = file
			break
		}
	}

	if assetName == "" {
		return fmt.Errorf("无法确定文件名")
	}

	assetPath := filepath.Join(tmpDir, assetName)

	if osName == "windows" {
		return u.installFromZip(assetPath, tmpDir, targetBinary)
	}

	return u.installFromTarGz(assetPath, tmpDir, targetBinary)
}

func (u *Updater) installFromZip(zipPath, tmpDir, targetBinary string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".exe") || !strings.Contains(f.Name, "/") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			binaryPath := filepath.Join(tmpDir, "cloud189.exe")
			out, err := os.Create(binaryPath)
			if err != nil {
				rc.Close()
				return err
			}

			_, err = io.Copy(out, rc)
			rc.Close()
			out.Close()

			if err != nil {
				return err
			}

			if err := os.Chmod(binaryPath, 0755); err != nil {
				return err
			}

			return replaceBinary(binaryPath, targetBinary)
		}
	}

	return fmt.Errorf("未找到二进制文件")
}

func (u *Updater) installFromTarGz(tarPath, tmpDir, targetBinary string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeReg && strings.HasPrefix(hdr.Name, "cloud189-") {
			binaryPath := filepath.Join(tmpDir, "cloud189")
			out, err := os.Create(binaryPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(out, tr)
			out.Close()

			if err != nil {
				return err
			}

			if err := os.Chmod(binaryPath, 0755); err != nil {
				return err
			}

			return replaceBinary(binaryPath, targetBinary)
		}
	}

	return fmt.Errorf("未找到二进制文件")
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	sourceInfo, err := source.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func replaceBinary(newBinary, targetBinary string) error {
	if runtime.GOOS == "windows" {
		return replaceBinaryWindows(newBinary, targetBinary)
	}
	return replaceBinaryUnix(newBinary, targetBinary)
}

func replaceBinaryWindows(newBinary, targetBinary string) error {
	oldPath := targetBinary + ".old"

	if err := os.Rename(targetBinary, oldPath); err != nil {
		return fmt.Errorf("无法重命名旧文件: %w", err)
	}

	if err := copyFile(newBinary, targetBinary); err != nil {
		os.Rename(oldPath, targetBinary)
		return fmt.Errorf("无法复制新文件: %w", err)
	}

	os.Remove(oldPath)

	fmt.Println("注意：Windows 平台更新后，旧版本文件将在重启后自动清理")
	return nil
}

func replaceBinaryUnix(newBinary, targetBinary string) error {
	if err := os.Rename(newBinary, targetBinary); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("权限不足，请使用 sudo 运行或确保有写入权限")
		}
		return fmt.Errorf("无法替换文件: %w", err)
	}
	return nil
}
