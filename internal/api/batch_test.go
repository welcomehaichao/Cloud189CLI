package api

import (
	"testing"
)

func TestBatchTaskInfoFields(t *testing.T) {
	task := BatchTaskInfo{
		FileId:   "file-123",
		FileName: "test.txt",
		IsFolder: 0,
	}

	if task.FileId != "file-123" {
		t.Errorf("FileId = %s, want file-123", task.FileId)
	}

	if task.FileName != "test.txt" {
		t.Errorf("FileName = %s, want test.txt", task.FileName)
	}
}

func TestBatchTaskResponseFields(t *testing.T) {
	resp := BatchTaskResponse{
		TaskID: "task-456",
	}

	if resp.TaskID != "task-456" {
		t.Errorf("TaskID = %s, want task-456", resp.TaskID)
	}
}

func TestUploadProgressMarkPartUploaded(t *testing.T) {
	progress := &UploadProgress{
		UploadParts: []string{"part1", "part2", "", "part4"},
	}

	progress.MarkPartUploaded(2)

	if progress.UploadParts[2] != "" {
		t.Error("MarkPartUploaded() should clear the part at index")
	}
}

func TestUploadProgressGetRemainingParts(t *testing.T) {
	tests := []struct {
		name     string
		progress *UploadProgress
		expected int
	}{
		{
			name: "all parts remaining",
			progress: &UploadProgress{
				UploadParts: []string{"p1", "p2", "p3"},
			},
			expected: 3,
		},
		{
			name: "some parts uploaded",
			progress: &UploadProgress{
				UploadParts: []string{"", "p2", "", "p4"},
			},
			expected: 2,
		},
		{
			name: "no parts remaining",
			progress: &UploadProgress{
				UploadParts: []string{"", "", ""},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := tt.progress.GetRemainingParts()
			if len(remaining) != tt.expected {
				t.Errorf("GetRemainingParts() = %d, want %d", len(remaining), tt.expected)
			}
		})
	}
}

func TestUploadProgressFields(t *testing.T) {
	progress := &UploadProgress{
		UploadFileId:   "upload-123",
		UploadFileSize: 1024000,
		SliceSize:      10240,
		PartInfos:      []string{"part1", "part2"},
		FileMd5:        "abc123",
		SliceMd5:       "def456",
		UploadParts:    []string{"p1", "p2"},
		LocalPath:      "/path/to/file",
	}

	if progress.UploadFileId != "upload-123" {
		t.Error("UploadFileId mismatch")
	}

	if progress.UploadFileSize != 1024000 {
		t.Error("UploadFileSize mismatch")
	}

	if progress.SliceSize != 10240 {
		t.Error("SliceSize mismatch")
	}
}
