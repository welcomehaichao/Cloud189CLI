package api

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
	"github.com/yuhaichao/cloud189-cli/pkg/utils"
)

type FilesResponse struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	FileListAO struct {
		Count      int          `json:"count"`
		FileList   []FileItem   `json:"fileList"`
		FolderList []FolderItem `json:"folderList"`
	} `json:"fileListAO"`
}

type FileItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	MD5        string `json:"md5"`
	LastOpTime string `json:"lastOpTime"`
	CreateDate string `json:"createDate"`
	Icon       struct {
		SmallURL  string `json:"smallUrl"`
		LargeURL  string `json:"largeUrl"`
		Max600    string `json:"max600"`
		MediumURL string `json:"mediumUrl"`
	} `json:"icon"`
}

type FolderItem struct {
	ID         int64  `json:"id"`
	ParentID   int64  `json:"parentId"`
	Name       string `json:"name"`
	LastOpTime string `json:"lastOpTime"`
	CreateDate string `json:"createDate"`
}

func (c *Client) ListFiles(folderId string, pageNum, pageSize int, orderBy, orderDirection string, isFamily bool) ([]types.File, error) {
	url := APIURL + "/listFiles.action"
	if isFamily {
		url = APIURL + "/family/file/listFiles.action"
	}

	var resp FilesResponse
	data, err := c.Get(url, func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"folderId":   folderId,
			"fileType":   "0",
			"mediaAttr":  "0",
			"iconOption": "5",
			"pageNum":    fmt.Sprint(pageNum),
			"pageSize":   fmt.Sprint(pageSize),
		})

		if isFamily {
			r.SetQueryParams(map[string]string{
				"familyId":   c.config.FamilyID,
				"orderBy":    toFamilyOrderBy(orderBy),
				"descending": toDesc(orderDirection),
			})
		} else {
			r.SetQueryParams(map[string]string{
				"recursive":  "0",
				"orderBy":    orderBy,
				"descending": toDesc(orderDirection),
			})
		}
	}, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	files := make([]types.File, 0, len(resp.FileListAO.FileList)+len(resp.FileListAO.FolderList))

	for _, folder := range resp.FileListAO.FolderList {
		modified, _ := utils.ParseTime(folder.LastOpTime)
		created, _ := utils.ParseTime(folder.CreateDate)
		files = append(files, types.File{
			ID:       strconv.FormatInt(folder.ID, 10),
			Name:     folder.Name,
			IsDir:    true,
			Modified: modified,
			Created:  created,
			ParentID: folder.ParentID,
		})
	}

	for _, file := range resp.FileListAO.FileList {
		modified, _ := utils.ParseTime(file.LastOpTime)
		created, _ := utils.ParseTime(file.CreateDate)
		files = append(files, types.File{
			ID:       strconv.FormatInt(file.ID, 10),
			Name:     file.Name,
			Size:     file.Size,
			IsDir:    false,
			MD5:      file.MD5,
			Modified: modified,
			Created:  created,
			Icon: types.Icon{
				SmallURL:  file.Icon.SmallURL,
				LargeURL:  file.Icon.LargeURL,
				Max600:    file.Icon.Max600,
				MediumURL: file.Icon.MediumURL,
			},
		})
	}

	return files, nil
}

func (c *Client) CreateFolder(parentFolderId, folderName string, isFamily bool) (*types.Folder, error) {
	url := APIURL + "/createFolder.action"
	if isFamily {
		url = APIURL + "/family/file/createFolder.action"
	}

	var folder FolderItem
	data, err := c.Post(url, func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"folderName":   folderName,
			"relativePath": "",
		})

		if isFamily {
			r.SetQueryParams(map[string]string{
				"familyId": c.config.FamilyID,
				"parentId": parentFolderId,
			})
		} else {
			r.SetQueryParam("parentFolderId", parentFolderId)
		}
	}, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &folder); err != nil {
		return nil, err
	}

	return &types.Folder{
		ID:       strconv.FormatInt(folder.ID, 10),
		Name:     folder.Name,
		ParentID: folder.ParentID,
	}, nil
}

func (c *Client) GetDownloadURL(fileId string, isFamily bool) (string, error) {
	url := APIURL + "/getFileDownloadUrl.action"
	if isFamily {
		url = APIURL + "/family/file/getFileDownloadUrl.action"
	}

	var resp struct {
		URL string `json:"fileDownloadUrl"`
	}

	data, err := c.Get(url, func(r *resty.Request) {
		r.SetQueryParam("fileId", fileId)

		if isFamily {
			r.SetQueryParam("familyId", c.config.FamilyID)
		} else {
			r.SetQueryParams(map[string]string{
				"dt":   "3",
				"flag": "1",
			})
		}
	}, isFamily)

	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	downloadURL := strings.ReplaceAll(resp.URL, "&amp;", "&")
	if !strings.HasPrefix(downloadURL, "http") {
		downloadURL = "https:" + downloadURL
	}

	client := resty.New().SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)

	res, err := client.R().Get(downloadURL)
	if err != nil {
		return "", err
	}

	if res.StatusCode() == 302 {
		downloadURL = res.Header().Get("Location")
	}

	return downloadURL, nil
}

type DownloadURLInfo struct {
	FileID      string    `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MD5         string    `json:"md5"`
	DownloadURL string    `json:"download_url"`
	ExpireTime  time.Time `json:"expire_time"`
	Expired     bool      `json:"expired"`
}

func parseExpireTime(downloadURL string) (time.Time, error) {
	// 尝试解析 X-Amz-Expires 格式（相对时间，秒数）
	reExpires := regexp.MustCompile(`(?i)X-Amz-Expires=([0-9]+)`)
	matchesExpires := reExpires.FindStringSubmatch(downloadURL)

	if len(matchesExpires) >= 2 {
		seconds, err := strconv.ParseInt(matchesExpires[1], 10, 64)
		if err == nil {
			// 相对于当前时间的过期时间
			expireTime := time.Now().Add(time.Duration(seconds) * time.Second)
			return expireTime, nil
		}
	}

	// 尝试解析 expire 或 Expires 格式（绝对时间戳）
	reExpire := regexp.MustCompile(`(?i)(?:expire|Expires)=([0-9]+)`)
	matchesExpire := reExpire.FindStringSubmatch(downloadURL)

	if len(matchesExpire) >= 2 {
		timestamp, err := strconv.ParseInt(matchesExpire[1], 10, 64)
		if err == nil {
			return time.Unix(timestamp, 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("expire time not found in URL")
}

func (c *Client) GetDownloadURLInfo(fileId string, isFamily bool) (*DownloadURLInfo, error) {
	downloadURL, err := c.GetDownloadURL(fileId, isFamily)
	if err != nil {
		return nil, err
	}

	expireTime, err := parseExpireTime(downloadURL)
	if err != nil {
		expireTime = time.Time{}
	}

	return &DownloadURLInfo{
		FileID:      fileId,
		DownloadURL: downloadURL,
		ExpireTime:  expireTime,
		Expired:     time.Now().After(expireTime) && !expireTime.IsZero(),
	}, nil
}

func (c *Client) RenameFile(fileId, newName string, isFamily bool) error {
	url := APIURL + "/renameFile.action"
	if isFamily {
		url = APIURL + "/family/file/renameFile.action"
	}

	_, err := c.Request(url, http.MethodPost, func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"fileId":       fileId,
			"destFileName": newName,
		})

		if isFamily {
			r.SetQueryParam("familyId", c.config.FamilyID)
		}
	}, isFamily)

	return err
}

func (c *Client) RenameFolder(folderId, newName string, isFamily bool) error {
	url := APIURL + "/renameFolder.action"
	if isFamily {
		url = APIURL + "/family/file/renameFolder.action"
	}

	_, err := c.Request(url, http.MethodPost, func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"folderId":       folderId,
			"destFolderName": newName,
		})

		if isFamily {
			r.SetQueryParam("familyId", c.config.FamilyID)
		}
	}, isFamily)

	return err
}

func toFamilyOrderBy(orderBy string) string {
	switch orderBy {
	case "filename":
		return "1"
	case "filesize":
		return "2"
	case "lastOpTime":
		return "3"
	default:
		return "1"
	}
}

func toDesc(orderDirection string) string {
	switch orderDirection {
	case "desc":
		return "true"
	default:
		return "false"
	}
}

type CreateUploadFileResp struct {
	UploadFileId   int64  `json:"uploadFileId"`
	FileUploadUrl  string `json:"fileUploadUrl"`
	FileCommitUrl  string `json:"fileCommitUrl"`
	FileDataExists int    `json:"fileDataExists"`
}

type OldCommitUploadFileResp struct {
	XMLName    xml.Name `xml:"file" json:"-"`
	ID         int64    `xml:"id" json:"id"`
	Name       string   `xml:"name" json:"name"`
	Size       int64    `xml:"size" json:"size"`
	MD5        string   `xml:"md5" json:"md5"`
	CreateDate string   `xml:"createDate" json:"createDate"`
}

type InitMultiUploadResp struct {
	Data struct {
		UploadType     int    `json:"uploadType"`
		UploadHost     string `json:"uploadHost"`
		UploadFileID   string `json:"uploadFileId"`
		FileDataExists int    `json:"fileDataExists"`
	} `json:"data"`
}

type UploadUrlsResp struct {
	Code string                    `json:"code"`
	Data map[string]UploadUrlsData `json:"uploadUrls"`
}

type UploadUrlsData struct {
	RequestURL    string `json:"requestURL"`
	RequestHeader string `json:"requestHeader"`
}

type UploadUrlInfo struct {
	PartNumber int
	Headers    map[string]string
	UploadUrlsData
}

type CommitMultiUploadFileResp struct {
	File struct {
		UserFileID string `json:"userFileId"`
		FileName   string `json:"fileName"`
		FileSize   int64  `json:"fileSize"`
		FileMd5    string `json:"fileMd5"`
		CreateDate string `json:"createDate"`
	} `json:"file"`
}

func (c *Client) OldUploadCreate(ctx context.Context, parentFolderId, fileMd5, fileName, fileSize string, isFamily bool) (*CreateUploadFileResp, error) {
	baseUrl := APIURL + "/createUploadFile.action"
	if isFamily {
		baseUrl = APIURL + "/family/file/createFamilyFile.action"
	}

	var resp CreateUploadFileResp
	data, err := c.Post(baseUrl, func(r *resty.Request) {
		r.SetContext(ctx)

		if isFamily {
			r.SetQueryParams(map[string]string{
				"familyId":     c.config.FamilyID,
				"parentId":     parentFolderId,
				"fileMd5":      fileMd5,
				"fileName":     fileName,
				"fileSize":     fileSize,
				"resumePolicy": "1",
			})
		} else {
			r.SetFormData(map[string]string{
				"parentFolderId": parentFolderId,
				"fileName":       fileName,
				"size":           fileSize,
				"md5":            fileMd5,
				"opertype":       "3",
				"flag":           "1",
				"resumePolicy":   "1",
				"isLog":          "0",
			})
		}
	}, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) InitMultiUpload(ctx context.Context, parentFolderId, fileName, fileSize, sliceSize string, isFamily bool) (*InitMultiUploadResp, error) {
	fullUrl := UploadURL
	if isFamily {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}

	params := Params{
		"parentFolderId": parentFolderId,
		"fileName":       fileName,
		"fileSize":       fileSize,
		"sliceSize":      sliceSize,
		"lazyCheck":      "1",
	}

	if isFamily {
		params.Set("familyId", c.config.FamilyID)
	}

	var resp InitMultiUploadResp
	data, err := c.RequestWithParams(fullUrl+"/initMultiUpload", http.MethodGet, func(r *resty.Request) {
		r.SetContext(ctx)
	}, params, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetMultiUploadUrls(ctx context.Context, uploadFileId string, partInfos []string, isFamily bool) ([]UploadUrlInfo, error) {
	fullUrl := UploadURL
	if isFamily {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}

	params := Params{
		"uploadFileId": uploadFileId,
		"partInfo":     strings.Join(partInfos, ","),
	}

	var resp struct {
		Code       string                    `json:"code"`
		Message    string                    `json:"message"`
		UploadUrls map[string]UploadUrlsData `json:"uploadUrls"`
	}

	data, err := c.RequestWithParams(fullUrl+"/getMultiUploadUrls", http.MethodGet, func(r *resty.Request) {
		r.SetContext(ctx)
	}, params, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v, data: %s", err, string(data))
	}

	if resp.Code != "" && resp.Code != "SUCCESS" {
		return nil, fmt.Errorf("API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	uploadUrls := resp.UploadUrls
	if uploadUrls == nil {
		uploadUrls = make(map[string]UploadUrlsData)
	}

	if len(uploadUrls) != len(partInfos) {
		return nil, fmt.Errorf("uploadUrls length mismatch: expected %d, got %d, data: %s", len(partInfos), len(uploadUrls), string(data))
	}

	uploadUrlInfos := make([]UploadUrlInfo, 0, len(uploadUrls))
	for k, uploadUrl := range uploadUrls {
		partNumber, err := strconv.Atoi(strings.TrimPrefix(k, "partNumber_"))
		if err != nil {
			return nil, err
		}
		uploadUrlInfos = append(uploadUrlInfos, UploadUrlInfo{
			PartNumber:     partNumber,
			Headers:        parseHttpHeader(uploadUrl.RequestHeader),
			UploadUrlsData: uploadUrl,
		})
	}

	sort.Slice(uploadUrlInfos, func(i, j int) bool {
		return uploadUrlInfos[i].PartNumber < uploadUrlInfos[j].PartNumber
	})

	return uploadUrlInfos, nil
}

func parseHttpHeader(str string) map[string]string {
	header := make(map[string]string)
	for _, value := range strings.Split(str, "&") {
		if k, v, found := strings.Cut(value, "="); found {
			header[k] = v
		}
	}
	return header
}

func (c *Client) CommitMultiUploadFile(ctx context.Context, uploadFileId, fileMd5, sliceMd5 string, isFamily bool) (*types.File, error) {
	fullUrl := UploadURL
	if isFamily {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}

	params := Params{
		"uploadFileId": uploadFileId,
		"fileMd5":      fileMd5,
		"sliceMd5":     sliceMd5,
		"lazyCheck":    "1",
		"isLog":        "0",
		"opertype":     "1",
	}

	var resp CommitMultiUploadFileResp
	data, err := c.RequestWithParams(fullUrl+"/commitMultiUploadFile", http.MethodGet, func(r *resty.Request) {
		r.SetContext(ctx)
	}, params, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	created, _ := utils.ParseTime(resp.File.CreateDate)
	return &types.File{
		ID:       resp.File.UserFileID,
		Name:     resp.File.FileName,
		Size:     resp.File.FileSize,
		MD5:      resp.File.FileMd5,
		IsDir:    false,
		Created:  created,
		Modified: created,
	}, nil
}

func calculatePartSize(fileSize int64) int64 {
	const DEFAULT = 10 * 1024 * 1024 // 10MB
	if fileSize > DEFAULT*2*999 {
		// 动态计算分片大小，确保不超过1999片
		partSize := int64(fileSize / 1999)
		// 向上取整到DEFAULT的整数倍，至少5倍（50MB）
		multiplier := (partSize / DEFAULT)
		if multiplier < 5 {
			multiplier = 5
		}
		return multiplier * DEFAULT
	}
	if fileSize > DEFAULT*999 {
		return DEFAULT * 2 // 20MB
	}
	return DEFAULT // 10MB
}

func (c *Client) StreamUpload(ctx context.Context, localPath, cloudPath string, progressCallback func(percent float64), isFamily bool) (*types.File, error) {
	return c.StreamUploadWithResume(ctx, localPath, cloudPath, progressCallback, isFamily, false)
}

func (c *Client) StreamUploadWithResume(ctx context.Context, localPath, cloudPath string, progressCallback func(percent float64), isFamily, resume bool) (*types.File, error) {
	fileInfo, err := utils.GetFileInfo(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	resolver := NewPathResolver(c, isFamily)
	parentFolderId, err := resolver.ResolvePath(cloudPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cloud path: %v", err)
	}

	fileSize := fileInfo.Size
	fileName := url.QueryEscape(fileInfo.Name)
	fileMd5Hex := fileInfo.MD5
	sliceSize := calculatePartSize(fileSize)

	count := 1
	if fileSize > sliceSize {
		count = int((fileSize + sliceSize - 1) / sliceSize)
	}

	progressManager := NewUploadProgressManager()
	sessionKey := c.config.SessionKey

	var uploadProgress *UploadProgress
	var initResp *InitMultiUploadResp

	if resume {
		uploadProgress, err = progressManager.LoadProgress(sessionKey, fileMd5Hex)
		if err != nil {
			fmt.Printf("Warning: failed to load progress: %v\n", err)
		}
	}

	if uploadProgress != nil && uploadProgress.UploadFileId != "" {
		initResp = &InitMultiUploadResp{
			Data: struct {
				UploadType     int    `json:"uploadType"`
				UploadHost     string `json:"uploadHost"`
				UploadFileID   string `json:"uploadFileId"`
				FileDataExists int    `json:"fileDataExists"`
			}{
				UploadFileID:   uploadProgress.UploadFileId,
				FileDataExists: 0,
			},
		}
		fmt.Printf("恢复上传进度：已上传 %d/%d 分片\n",
			count-len(uploadProgress.GetRemainingParts()), count)
	} else {
		initResp, err = c.InitMultiUpload(ctx, parentFolderId, fileName,
			strconv.FormatInt(fileSize, 10), strconv.FormatInt(sliceSize, 10), isFamily)
		if err != nil {
			return nil, fmt.Errorf("failed to init multi upload: %v", err)
		}

		if initResp.Data.FileDataExists == 1 {
			if progressCallback != nil {
				progressCallback(100.0)
			}
			return c.CommitMultiUploadFile(ctx, initResp.Data.UploadFileID, fileMd5Hex, fileMd5Hex, isFamily)
		}

		file, err := utils.OpenFile(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		sliceMd5Hexs := make([]string, 0, count)
		partInfos := make([]string, 0, count)

		for i := 1; i <= count; i++ {
			offset := int64(i-1) * sliceSize
			partSize := sliceSize
			if i == count {
				partSize = fileSize - offset
			}

			section := io.NewSectionReader(file, offset, partSize)
			md5Hash, err := utils.MD5HashReaderUpper(section)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate part %d md5: %v", i, err)
			}

			sliceMd5Hexs = append(sliceMd5Hexs, md5Hash)
			partInfo := fmt.Sprintf("%d-%s", i, base64EncodeHex(md5Hash))
			partInfos = append(partInfos, partInfo)
		}

		sliceMd5Hex := fileMd5Hex
		if fileSize > sliceSize {
			sliceMd5Hex = strings.ToUpper(utils.MD5Hash([]byte(strings.Join(sliceMd5Hexs, "\n"))))
		}

		uploadProgress = &UploadProgress{
			UploadFileId:   initResp.Data.UploadFileID,
			UploadFileSize: fileSize,
			SliceSize:      sliceSize,
			PartInfos:      partInfos,
			FileMd5:        fileMd5Hex,
			SliceMd5:       sliceMd5Hex,
			UploadParts:    partInfos,
			LocalPath:      localPath,
		}

		if resume {
			if err := progressManager.SaveProgress(sessionKey, uploadProgress); err != nil {
				fmt.Printf("Warning: failed to save progress: %v\n", err)
			}
		}
	}

	file, err := utils.OpenFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	uploadedCount := 0
	for i, partInfo := range uploadProgress.UploadParts {
		if ctx.Err() != nil {
			if resume && ctx.Err() == context.Canceled {
				if err := progressManager.SaveProgress(sessionKey, uploadProgress); err != nil {
					fmt.Printf("Warning: failed to save progress on cancel: %v\n", err)
				} else {
					fmt.Printf("\n上传已取消，进度已保存。下次可使用 --resume 继续\n")
				}
			}
			return nil, ctx.Err()
		}

		if partInfo == "" {
			uploadedCount++
			continue
		}

		offset := int64(i) * uploadProgress.SliceSize
		partSize := uploadProgress.SliceSize
		if i == count-1 {
			partSize = uploadProgress.UploadFileSize - offset
		}

		uploadUrls, err := c.GetMultiUploadUrls(ctx, uploadProgress.UploadFileId, []string{partInfo}, isFamily)
		if err != nil {
			return nil, fmt.Errorf("failed to get upload url for part %d: %v", i+1, err)
		}

		uploadUrl := uploadUrls[0]
		// Directly use RequestURL without modification - the URL should be complete
		uploadURL := uploadUrl.RequestURL
		// OpenList doesn't add "https:" prefix, the URL is already complete from API

		section := io.NewSectionReader(file, offset, partSize)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, section)
		if err != nil {
			return nil, err
		}

		query := req.URL.Query()
		for key, value := range c.ClientSuffix() {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()

		for key, value := range uploadUrl.Headers {
			req.Header.Add(key, value)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			if resume {
				progressManager.SaveProgress(sessionKey, uploadProgress)
			}
			return nil, fmt.Errorf("failed to upload part %d: %v", i+1, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if resume {
				progressManager.SaveProgress(sessionKey, uploadProgress)
			}
			return nil, fmt.Errorf("upload part %d failed with status %d: %s", i+1, resp.StatusCode, string(body))
		}

		uploadProgress.MarkPartUploaded(i)
		uploadedCount++

		if progressCallback != nil {
			progress := float64(uploadedCount) * 100.0 / float64(count)
			progressCallback(progress)
		}

		if resume && i%5 == 0 {
			progressManager.SaveProgress(sessionKey, uploadProgress)
		}
	}

	result, err := c.CommitMultiUploadFile(ctx, uploadProgress.UploadFileId,
		uploadProgress.FileMd5, uploadProgress.SliceMd5, isFamily)

	if err == nil {
		progressManager.DeleteProgress(sessionKey, uploadProgress.FileMd5)
		if progressCallback != nil {
			progressCallback(100.0)
		}
	}

	return result, err
}

func base64EncodeHex(hexStr string) string {
	data, _ := hex.DecodeString(hexStr)
	return base64.StdEncoding.EncodeToString(data)
}

type ShareInfo struct {
	ShareId    string `json:"shareId"`
	ShareLink  string `json:"shareLink"`
	AccessCode string `json:"accessCode"`
	ExpireTime string `json:"expireTime"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	FileId     string `json:"fileId"`
	IsFolder   int    `json:"isFolder"`
}

type CreateShareLinkResp struct {
	ResCode       int    `json:"res_code"`
	ResMessage    string `json:"res_message"`
	ShareLinkList []struct {
		ShareId    int64  `json:"shareId"`
		ShareLink  string `json:"accessUrl"`
		AccessCode string `json:"accessCode"`
		FileId     int64  `json:"fileId"`
		URL        string `json:"url"`
	} `json:"shareLinkList"`
}

func (c *Client) CreateShareLink(fileId string, isFolder bool, expireDays int, accessCode string, isFamily bool) (*ShareInfo, error) {
	url := APIURL + "/createShareLink.action"
	if isFamily {
		url = APIURL + "/family/file/createShareLink.action"
	}

	isFolderInt := 0
	if isFolder {
		isFolderInt = 1
	}

	params := map[string]string{
		"fileId":   fileId,
		"isFolder": strconv.Itoa(isFolderInt),
	}

	if expireDays > 0 {
		params["expireDays"] = strconv.Itoa(expireDays)
	} else {
		params["expireDays"] = "0"
	}

	if accessCode != "" {
		params["accessCode"] = accessCode
		params["needAccessCode"] = "1"
	} else {
		params["needAccessCode"] = "0"
	}

	if isFamily {
		params["familyId"] = c.config.FamilyID
	}

	var resp CreateShareLinkResp
	data, err := c.Post(url, func(r *resty.Request) {
		r.SetQueryParams(params)
	}, isFamily)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.ResCode != 0 {
		return nil, fmt.Errorf("create share link failed: %s", resp.ResMessage)
	}

	if len(resp.ShareLinkList) == 0 {
		return nil, fmt.Errorf("create share link failed: no share link returned")
	}

	shareLinkInfo := resp.ShareLinkList[0]
	return &ShareInfo{
		ShareId:    strconv.FormatInt(shareLinkInfo.ShareId, 10),
		ShareLink:  shareLinkInfo.ShareLink,
		AccessCode: shareLinkInfo.AccessCode,
		FileId:     fileId,
		IsFolder:   isFolderInt,
	}, nil
}

func (c *Client) CancelShare(shareId string) error {
	url := APIURL + "/cancelShare.action"

	data, err := c.Post(url, func(r *resty.Request) {
		r.SetQueryParam("shareId", shareId)
	}, false)

	if err != nil {
		return err
	}

	var resp struct {
		ResCode    int    `json:"res_code"`
		ResMessage string `json:"res_message"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	if resp.ResCode != 0 {
		return fmt.Errorf("cancel share failed: %s", resp.ResMessage)
	}

	return nil
}

func (c *Client) DownloadFile(ctx context.Context, fileId, localPath string, progressCallback func(percent float64), isFamily bool) error {
	downloadURL, err := c.GetDownloadURL(fileId, isFamily)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %v", err)
	}

	return c.DownloadFromURL(ctx, downloadURL, localPath, progressCallback)
}

func (c *Client) DownloadFromURL(ctx context.Context, downloadURL, localPath string, progressCallback func(percent float64)) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	fileSize := resp.ContentLength
	if fileSize <= 0 {
		fileSize = 0
	}

	var downloaded int64 = 0
	buf := make([]byte, 32*1024)

	for {
		n, err := io.CopyBuffer(file, resp.Body, buf)
		downloaded += n

		if progressCallback != nil && fileSize > 0 {
			progress := float64(downloaded) * 100.0 / float64(fileSize)
			progressCallback(progress)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("download error: %v", err)
		}

		if n == 0 {
			break
		}
	}

	if progressCallback != nil {
		progressCallback(100.0)
	}

	return nil
}

func (c *Client) DownloadFileWithResume(ctx context.Context, fileId, localPath string, progressCallback func(percent float64), isFamily bool) error {
	downloadURL, err := c.GetDownloadURL(fileId, isFamily)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %v", err)
	}

	return c.DownloadWithResume(ctx, downloadURL, localPath, progressCallback)
}

func (c *Client) DownloadWithResume(ctx context.Context, downloadURL, localPath string, progressCallback func(percent float64)) error {
	var existingSize int64 = 0

	if info, err := os.Stat(localPath); err == nil {
		existingSize = info.Size()
	}

	file, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	if existingSize > 0 {
		if _, err := file.Seek(existingSize, 0); err != nil {
			return fmt.Errorf("failed to seek file: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download: %v", err)
	}
	defer resp.Body.Close()

	var statusCodeOK bool
	if existingSize > 0 {
		statusCodeOK = resp.StatusCode == http.StatusPartialContent
	} else {
		statusCodeOK = resp.StatusCode == http.StatusOK
	}

	if !statusCodeOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	if existingSize > 0 {
		totalSize += existingSize
	}

	var downloaded int64 = existingSize
	buf := make([]byte, 32*1024)

	for {
		n, err := io.CopyBuffer(file, resp.Body, buf)
		downloaded += n

		if progressCallback != nil && totalSize > 0 {
			progress := float64(downloaded) * 100.0 / float64(totalSize)
			progressCallback(progress)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("download error: %v", err)
		}

		if n == 0 {
			break
		}
	}

	if progressCallback != nil {
		progressCallback(100.0)
	}

	return nil
}

func (c *Client) OldUploadCommit(ctx context.Context, commitUrl, uploadFileId string, isFamily bool) (*types.File, error) {
	var resp OldCommitUploadFileResp

	data, err := c.Post(commitUrl, func(r *resty.Request) {
		r.SetContext(ctx)

		if isFamily {
			r.SetHeaders(map[string]string{
				"ResumePolicy": "1",
				"UploadFileId": uploadFileId,
				"FamilyId":     c.config.FamilyID,
			})
		} else {
			r.SetFormData(map[string]string{
				"opertype":     "1",
				"resumePolicy": "1",
				"uploadFileId": uploadFileId,
				"isLog":        "0",
			})
		}
	}, isFamily)

	if err != nil {
		return nil, err
	}

	// Try XML first
	if err := xml.Unmarshal(data, &resp); err != nil {
		// If XML fails, try JSON
		if jsonErr := json.Unmarshal(data, &resp); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse response as XML or JSON: xml_error=%v, json_error=%v", err, jsonErr)
		}
	}

	created, _ := utils.ParseTime(resp.CreateDate)
	return &types.File{
		ID:       strconv.FormatInt(resp.ID, 10),
		Name:     resp.Name,
		Size:     resp.Size,
		MD5:      resp.MD5,
		IsDir:    false,
		Created:  created,
		Modified: created,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, localPath, cloudPath string, progressCallback func(percent float64), isFamily bool) (*types.File, error) {
	fileInfo, err := utils.GetFileInfo(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	resolver := NewPathResolver(c, isFamily)
	parentFolderId, err := resolver.ResolvePath(cloudPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cloud path: %v", err)
	}

	fileMd5 := fileInfo.MD5
	fileName := url.QueryEscape(fileInfo.Name)
	fileSize := strconv.FormatInt(fileInfo.Size, 10)

	uploadInfo, err := c.OldUploadCreate(ctx, parentFolderId, fileMd5, fileName, fileSize, isFamily)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload session: %v", err)
	}

	if uploadInfo.FileDataExists == 1 {
		if progressCallback != nil {
			progressCallback(100.0)
		}
		return c.OldUploadCommit(ctx, uploadInfo.FileCommitUrl, strconv.FormatInt(uploadInfo.UploadFileId, 10), isFamily)
	}

	file, err := utils.OpenFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	uploadUrl := uploadInfo.FileUploadUrl
	if !strings.HasPrefix(uploadUrl, "http") {
		if strings.HasPrefix(uploadUrl, "//") {
			uploadUrl = "https:" + uploadUrl
		} else if strings.HasPrefix(uploadUrl, "/") {
			uploadUrl = "https://upload.cloud.189.cn" + uploadUrl
		} else if uploadUrl != "" {
			uploadUrl = "https://upload.cloud.189.cn/" + uploadUrl
		} else {
			return nil, fmt.Errorf("empty upload URL received from API")
		}
	}

	headers := map[string]string{
		"ResumePolicy": "1",
		"Expect":       "100-continue",
	}

	if isFamily {
		headers["FamilyId"] = c.config.FamilyID
		headers["UploadFileId"] = strconv.FormatInt(uploadInfo.UploadFileId, 10)
	} else {
		headers["Edrive-UploadFileId"] = strconv.FormatInt(uploadInfo.UploadFileId, 10)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadUrl, file)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, value := range c.ClientSuffix() {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	for key, value := range c.SignatureHeader(uploadUrl, http.MethodPut, isFamily) {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	if progressCallback != nil {
		progressCallback(100.0)
	}

	return c.OldUploadCommit(ctx, uploadInfo.FileCommitUrl, strconv.FormatInt(uploadInfo.UploadFileId, 10), isFamily)
}
