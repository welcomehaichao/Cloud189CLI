package api

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

type BatchTaskInfo struct {
	FileId      string `json:"fileId"`
	FileName    string `json:"fileName"`
	IsFolder    int    `json:"isFolder"`
	SrcParentId string `json:"srcParentId,omitempty"`
	DealWay     int    `json:"dealWay,omitempty"`
	IsConflict  int    `json:"isConflict,omitempty"`
}

type BatchTaskResponse struct {
	TaskID string `json:"taskId"`
}

type BatchTaskStateResponse struct {
	FailedCount         int     `json:"failedCount"`
	Process             int     `json:"process"`
	SkipCount           int     `json:"skipCount"`
	SubTaskCount        int     `json:"subTaskCount"`
	SuccessedCount      int     `json:"successedCount"`
	SuccessedFileIDList []int64 `json:"successedFileIdList"`
	TaskID              string  `json:"taskId"`
	TaskStatus          int     `json:"taskStatus"`
}

func (c *Client) CreateBatchTask(taskType, familyID, targetFolderId string, taskInfos []BatchTaskInfo) (*BatchTaskResponse, error) {
	taskInfosJSON, _ := json.Marshal(taskInfos)

	var resp BatchTaskResponse
	data, err := c.Post(APIURL+"/batch/createBatchTask.action", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"type":      taskType,
			"taskInfos": string(taskInfosJSON),
		})

		if targetFolderId != "" {
			r.SetFormData(map[string]string{"targetFolderId": targetFolderId})
		}

		if familyID != "" {
			r.SetFormData(map[string]string{"familyId": familyID})
		}
	}, familyID != "")

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) CheckBatchTask(taskType, taskID string) (*BatchTaskStateResponse, error) {
	var resp BatchTaskStateResponse
	data, err := c.Post(APIURL+"/batch/checkBatchTask.action", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"type":   taskType,
			"taskId": taskID,
		})
	}, false)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) WaitBatchTask(taskType, taskID string, interval time.Duration) error {
	for {
		state, err := c.CheckBatchTask(taskType, taskID)
		if err != nil {
			return err
		}

		switch state.TaskStatus {
		case 2:
			return types.NewCloudError("TASK_CONFLICT", "task has conflicts")
		case 4:
			return nil
		}

		time.Sleep(interval)
	}
}

func (c *Client) Move(srcObj, dstDir *types.File, isFamily bool) error {
	familyID := ""
	if isFamily {
		familyID = c.config.FamilyID
	}

	isFolder := 0
	if srcObj.IsDir {
		isFolder = 1
	}

	taskInfos := []BatchTaskInfo{
		{
			FileId:   srcObj.ID,
			FileName: srcObj.Name,
			IsFolder: isFolder,
		},
	}

	resp, err := c.CreateBatchTask("MOVE", familyID, dstDir.ID, taskInfos)
	if err != nil {
		return err
	}

	return c.WaitBatchTask("MOVE", resp.TaskID, 400*time.Millisecond)
}

func (c *Client) Copy(srcObj, dstDir *types.File, isFamily bool) error {
	familyID := ""
	if isFamily {
		familyID = c.config.FamilyID
	}

	isFolder := 0
	if srcObj.IsDir {
		isFolder = 1
	}

	taskInfos := []BatchTaskInfo{
		{
			FileId:   srcObj.ID,
			FileName: srcObj.Name,
			IsFolder: isFolder,
		},
	}

	resp, err := c.CreateBatchTask("COPY", familyID, dstDir.ID, taskInfos)
	if err != nil {
		return err
	}

	return c.WaitBatchTask("COPY", resp.TaskID, time.Second)
}

func (c *Client) Delete(obj *types.File, isFamily bool) error {
	familyID := ""
	if isFamily {
		familyID = c.config.FamilyID
	}

	isFolder := 0
	if obj.IsDir {
		isFolder = 1
	}

	taskInfos := []BatchTaskInfo{
		{
			FileId:   obj.ID,
			FileName: obj.Name,
			IsFolder: isFolder,
		},
	}

	resp, err := c.CreateBatchTask("DELETE", familyID, "", taskInfos)
	if err != nil {
		return err
	}

	return c.WaitBatchTask("DELETE", resp.TaskID, 200*time.Millisecond)
}

func (c *Client) ClearRecycle(familyID string, taskInfos []BatchTaskInfo) error {
	resp, err := c.CreateBatchTask("CLEAR_RECYCLE", familyID, "", taskInfos)
	if err != nil {
		return err
	}

	return c.WaitBatchTask("CLEAR_RECYCLE", resp.TaskID, time.Second)
}
