package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
)

func (c *Client) GetFamilyList() ([]types.FamilyInfo, error) {
	var resp types.FamilyInfoList
	data, err := c.Get(APIURL+"/family/manage/getFamilyList.action", nil, true)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.FamilyInfoResp, nil
}

func (c *Client) GetFamilyID() (string, error) {
	infos, err := c.GetFamilyList()
	if err != nil {
		return "", err
	}

	if len(infos) == 0 {
		return "", fmt.Errorf("no family found")
	}

	for _, info := range infos {
		if strings.Contains(c.config.Username, info.RemarkName) {
			return strconv.FormatInt(info.FamilyID, 10), nil
		}
	}

	return strconv.FormatInt(infos[0].FamilyID, 10), nil
}

func (c *Client) SaveFamilyFileToPersonCloud(familyID string, srcObj, dstDir *types.File, overwrite bool) error {
	isFolder := 0
	if srcObj.IsDir {
		isFolder = 1
	}

	task := BatchTaskInfo{
		FileId:   srcObj.ID,
		FileName: srcObj.Name,
		IsFolder: isFolder,
	}

	resp, err := c.CreateBatchTaskWithOther("COPY", familyID, dstDir.ID, map[string]string{
		"groupId":  "null",
		"copyType": "2",
		"shareId":  "null",
	}, task)

	if err != nil {
		return err
	}

	for {
		state, err := c.CheckBatchTask("COPY", resp.TaskID)
		if err != nil {
			return err
		}

		switch state.TaskStatus {
		case 2:
			task.DealWay = 3
			if !overwrite {
				task.DealWay = 2
			}
			if err := c.ManageBatchTask("COPY", resp.TaskID, dstDir.ID, task); err != nil {
				return err
			}
		case 4:
			return nil
		}

		time.Sleep(400 * time.Millisecond)
	}
}

func (c *Client) ManageBatchTask(taskType, taskID, targetFolderId string, taskInfos ...BatchTaskInfo) error {
	taskInfosJSON, _ := json.Marshal(taskInfos)

	_, err := c.Post(APIURL+"/batch/manageBatchTask.action", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"targetFolderId": targetFolderId,
			"type":           taskType,
			"taskId":         taskID,
			"taskInfos":      string(taskInfosJSON),
		})
	}, false)

	return err
}

func (c *Client) CreateBatchTaskWithOther(taskType, familyID, targetFolderId string, other map[string]string, taskInfos ...BatchTaskInfo) (*BatchTaskResponse, error) {
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

		for k, v := range other {
			r.SetFormData(map[string]string{k: v})
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

func (c *Client) GetCapacityInfo() (*CapacityResponse, error) {
	var resp CapacityResponse
	data, err := c.Get(APIURL+"/portal/getUserSizeInfo.action", nil, false)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capacity info: %w, data: %s", err, string(data))
	}

	if resp.ResCode != 0 {
		return nil, fmt.Errorf("API error: %s (code: %d)", resp.ResMessage, resp.ResCode)
	}

	return &resp, nil
}

type CapacityResponse struct {
	ResCode           int    `json:"res_code"`
	ResMessage        string `json:"res_message"`
	Account           string `json:"account"`
	CloudCapacityInfo struct {
		FreeSize     int64 `json:"freeSize"`
		MailUsedSize int64 `json:"mail189UsedSize"`
		TotalSize    int64 `json:"totalSize"`
		UsedSize     int64 `json:"usedSize"`
	} `json:"cloudCapacityInfo"`
	FamilyCapacityInfo struct {
		FreeSize  int64 `json:"freeSize"`
		TotalSize int64 `json:"totalSize"`
		UsedSize  int64 `json:"usedSize"`
	} `json:"familyCapacityInfo"`
	TotalSize uint64 `json:"totalSize"`
}
