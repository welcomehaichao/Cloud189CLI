package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name string
	Size int64
	Path string
	MD5  string
}

func GetFileInfo(path string) (*FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	md5Hash := hex.EncodeToString(hash.Sum(nil))
	md5Hash = strings.ToUpper(md5Hash)

	return &FileInfo{
		Name: stat.Name(),
		Size: stat.Size(),
		Path: path,
		MD5:  md5Hash,
	}, nil
}

func OpenFile(path string) (*os.File, error) {
	return os.Open(path)
}

func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cloud189")
}
