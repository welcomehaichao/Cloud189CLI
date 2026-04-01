package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuhaichao/cloud189-cli/pkg/utils"
)

type UploadProgress struct {
	UploadFileId   string   `json:"uploadFileId"`
	UploadFileSize int64    `json:"uploadFileSize"`
	SliceSize      int64    `json:"sliceSize"`
	PartInfos      []string `json:"partInfos"`
	FileMd5        string   `json:"fileMd5"`
	SliceMd5       string   `json:"sliceMd5"`
	UploadParts    []string `json:"uploadParts"`
	LocalPath      string   `json:"localPath"`
}

type UploadProgressManager struct {
	progressDir string
}

func NewUploadProgressManager() *UploadProgressManager {
	configDir := utils.GetConfigDir()
	progressDir := filepath.Join(configDir, "upload_progress")

	if err := os.MkdirAll(progressDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create progress directory: %v\n", err)
	}

	return &UploadProgressManager{
		progressDir: progressDir,
	}
}

func (m *UploadProgressManager) getProgressFilePath(sessionKey, fileMd5 string) string {
	progressFile := fmt.Sprintf("%s_%s.json",
		utils.MD5Hash([]byte(sessionKey)),
		fileMd5)
	return filepath.Join(m.progressDir, progressFile)
}

func (m *UploadProgressManager) SaveProgress(sessionKey string, progress *UploadProgress) error {
	progressFile := m.getProgressFilePath(sessionKey, progress.FileMd5)

	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %v", err)
	}

	if err := os.WriteFile(progressFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %v", err)
	}

	return nil
}

func (m *UploadProgressManager) LoadProgress(sessionKey, fileMd5 string) (*UploadProgress, error) {
	progressFile := m.getProgressFilePath(sessionKey, fileMd5)

	data, err := os.ReadFile(progressFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read progress file: %v", err)
	}

	var progress UploadProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal progress: %v", err)
	}

	return &progress, nil
}

func (m *UploadProgressManager) DeleteProgress(sessionKey, fileMd5 string) error {
	progressFile := m.getProgressFilePath(sessionKey, fileMd5)

	if err := os.Remove(progressFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete progress file: %v", err)
	}

	return nil
}

func (m *UploadProgressManager) CleanCompletedProgress() error {
	files, err := os.ReadDir(m.progressDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read progress directory: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			progressFile := filepath.Join(m.progressDir, file.Name())

			data, err := os.ReadFile(progressFile)
			if err != nil {
				continue
			}

			var progress UploadProgress
			if err := json.Unmarshal(data, &progress); err != nil {
				continue
			}

			allUploaded := true
			for _, part := range progress.UploadParts {
				if part != "" {
					allUploaded = false
					break
				}
			}

			if allUploaded {
				os.Remove(progressFile)
			}
		}
	}

	return nil
}

func (p *UploadProgress) GetRemainingParts() []string {
	remaining := make([]string, 0)
	for _, part := range p.UploadParts {
		if part != "" {
			remaining = append(remaining, part)
		}
	}
	return remaining
}

func (p *UploadProgress) IsCompleted() bool {
	for _, part := range p.UploadParts {
		if part != "" {
			return false
		}
	}
	return true
}

func (p *UploadProgress) MarkPartUploaded(index int) {
	if index >= 0 && index < len(p.UploadParts) {
		p.UploadParts[index] = ""
	}
}
