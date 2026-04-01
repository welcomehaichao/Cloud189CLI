# 天翼云盘 TV电视端 API 文档

## 一、概述

TV电视端API是天翼云盘专门为智能电视和机顶盒设计的接口，特点是通过二维码扫码登录，无需输入密码，适合电视大屏使用场景。

### 基本信息

| 项目 | 值 |
|------|-----|
| API基础URL | https://api.cloud.189.cn |
| AppKey | 600100885 |
| AppSignatureSecret | fe5734c74c2f96a38157f420b32dc995 |
| Version | 6.5.5 |
| ClientType | FAMILY_TV |
| ChannelId | home02 |
| User-Agent | EcloudTV/6.5.5 (PJX110; unknown; home02) Android/35 |

### 特点

- ✅ 二维码扫码登录，无需输入密码
- ✅ 支持个人云和家庭云
- ✅ 支持旧版上传
- ✅ 支持秒传
- ✅ 完整的家庭云操作支持
- ❌ 不支持新版上传（stream/rapid）
- ❌ 家庭云不支持覆盖上传

---

## 二、认证机制

### 2.1 二维码登录流程

TV端仅支持二维码扫码登录，无需输入密码。

#### 步骤1: 获取二维码UUID

```http
GET https://api.cloud.189.cn/family/manage/getQrCodeUUID.action
Accept: application/json;charset=UTF-8
```

**请求头：** AppKey签名
```
Timestamp: {毫秒时间戳}
X-Request-ID: {uuid}
AppKey: 600100885
AppSignature: {AppKey签名}
```

**必需URL参数：**
```
clientType: FAMILY_TV
version: 6.5.5
channelId: home02
clientSn: unknown
model: PJX110
osFamily: Android
osVersion: 35
networkAccessMode: WIFI
telecomsOperator: 46011
```

**响应示例：**
```json
{
  "uuid": "https://m.cloud.189.cn/zhuanti/qrLogin/qrCode/index.html?uuid={uuid}"
}
```

#### 步骤2: 生成并展示二维码

```javascript
// 从uuid提取二维码内容
qrContent = response.uuid  // 或提取uuid参数

// 生成二维码图片
qrCodeImage = QRCode.generate(qrContent, {
  size: 256,
  level: 'M'
})

// 展示给用户
displayQRCode(qrCodeImage)

// 提供扫码链接
console.log("扫描链接:", qrContent)
```

#### 步骤3: 轮询扫码状态

```http
GET https://api.cloud.189.cn/family/manage/qrcodeLoginResult.action
Accept: application/json;charset=UTF-8
```

**请求头：** AppKey签名
```
Timestamp: {毫秒时间戳}
X-Request-ID: {uuid}
AppKey: 600100885
AppSignature: {AppKey签名}
```

**请求参数：**
```
uuid: {从步骤1获取的uuid}
clientType: FAMILY_TV
version: 6.5.5
channelId: home02
...（其他客户端参数）
```

**响应示例：**
```json
{
  "accessToken": "e189AccessToken值",
  "expiresIn": 过期时间戳
}
```

**轮询逻辑：**
```javascript
async function waitForQRCodeScan(uuid) {
  const maxRetry = 60  // 最大重试60次（约5分钟）
  
  for (let i = 0; i < maxRetry; i++) {
    try {
      const response = await getQRCodeLoginResult(uuid)
      
      if (response.accessToken) {
        return response.accessToken
      }
    } catch (error) {
      // 继续轮询
    }
    
    await sleep(5000)  // 每5秒轮询一次
  }
  
  throw new Error('二维码已过期')
}
```

#### 步骤4: 获取Session

```http
GET https://api.cloud.189.cn/family/manage/loginFamilyMerge.action
Accept: application/json;charset=UTF-8
```

**请求头：** AppKey签名
```
Timestamp: {毫秒时间戳}
X-Request-ID: {uuid}
AppKey: 600100885
AppSignature: {AppKey签名}
```

**请求参数：**
```
e189AccessToken: {从步骤3获取的accessToken}
clientType: FAMILY_TV
version: 6.5.5
channelId: home02
...（其他客户端参数）
```

**响应示例：**
```json
{
  "res_code": 0,
  "res_message": "",
  "loginName": "用户名",
  "keepAlive": 300,
  "getFileDiffSpan": 30,
  "getUserInfoSpan": 120,
  "sessionKey": "{sessionKey}",
  "sessionSecret": "{sessionSecret}",
  "familySessionKey": "{familySessionKey}",
  "familySessionSecret": "{familySessionSecret}",
  "isSaveName": "",
  "accessToken": "{accessToken}",
  "refreshToken": "{refreshToken}"
}
```

**重要字段说明：**
- `sessionKey`: 个人云会话密钥
- `sessionSecret`: 个人云会话密钥（用于签名）
- `familySessionKey`: 家庭云会话密钥
- `familySessionSecret`: 家庭云会话密钥（用于签名）
- `accessToken`: 访问令牌
- `refreshToken`: 刷新令牌

### 2.2 Session保活

```http
GET https://api.cloud.189.cn/keepUserSession.action
Accept: application/json;charset=UTF-8
```

**请求头：** SessionKey签名
```
Date: {HTTP时间格式}
SessionKey: {sessionKey}
X-Request-ID: {uuid}
Signature: {SessionKey签名}
```

**必需URL参数：**
```
clientType: FAMILY_TV
version: 6.5.5
channelId: home02
...（其他客户端参数）
```

**建议频率：** 每5分钟执行一次

### 2.3 Session刷新

使用accessToken重新获取Session：

```http
GET https://api.cloud.189.cn/family/manage/loginFamilyMerge.action
```

**请求参数：**
```
e189AccessToken: {accessToken}
```

**注意：** 需要保存accessToken，Session失效时可以用它重新获取

---

## 三、签名机制

TV端使用两种签名机制：AppKey签名和SessionKey签名。

### 3.1 AppKey签名

用于登录相关的接口（获取二维码、检查扫码状态、获取Session）。

#### 签名算法

```javascript
function AppKeySignature(appSignatureSecret, appKey, method, url, timestamp) {
  // 1. 提取URI路径
  urlpath = url.match(/:\/\/[^\/]+((\/[^\/\s?#]+)*)/)[1]
  
  // 2. 构建签名字符串
  data = "AppKey=" + appKey
  data += "&Operate=" + method
  data += "&RequestURI=" + urlpath
  data += "&Timestamp=" + timestamp
  
  // 3. HMAC-SHA1签名
  signature = HMAC_SHA1(appSignatureSecret, data)
  
  // 4. Hex编码（大写）
  return HexEncode(signature).toUpperCase()
}
```

#### 签名示例

```javascript
const appSignatureSecret = "fe5734c74c2f96a38157f420b32dc995"
const appKey = "600100885"
const method = "GET"
const url = "https://api.cloud.189.cn/family/manage/getQrCodeUUID.action"
const timestamp = Date.now()

const signature = AppKeySignature(appSignatureSecret, appKey, method, url, timestamp)
// 结果: "A1B2C3D4E5F6..."
```

#### 请求头

```http
Timestamp: {毫秒时间戳}
X-Request-ID: {uuid}
AppKey: 600100885
AppSignature: {签名}
```

### 3.2 SessionKey签名

用于文件操作相关的接口。

#### 签名算法

```javascript
function SessionKeySignature(sessionSecret, sessionKey, method, url, date) {
  // 1. 提取URI路径
  urlpath = url.match(/:\/\/[^\/]+((\/[^\/\s?#]+)*)/)[1]
  
  // 2. 构建签名字符串
  data = "SessionKey=" + sessionKey
  data += "&Operate=" + method
  data += "&RequestURI=" + urlpath
  data += "&Date=" + date
  
  // 注意：TV端SessionKey签名不包含params
  
  // 3. HMAC-SHA1签名
  signature = HMAC_SHA1(sessionSecret, data)
  
  // 4. Hex编码（大写）
  return HexEncode(signature).toUpperCase()
}
```

#### 签名示例

```javascript
const sessionSecret = "abc123..."
const sessionKey = "xyz789..."
const method = "GET"
const url = "https://api.cloud.189.cn/listFiles.action"
const date = "Mon, 01 Jan 2024 12:00:00 GMT"

const signature = SessionKeySignature(sessionSecret, sessionKey, method, url, date)
// 结果: "A1B2C3D4E5F6..."
```

#### 请求头

```http
Accept: application/json;charset=UTF-8
Date: {HTTP时间格式}
SessionKey: {sessionKey或familySessionKey}
X-Request-ID: {uuid}
Signature: {签名}
```

**注意：** 
- 个人云操作使用 `sessionKey` 和 `sessionSecret`
- 家庭云操作使用 `familySessionKey` 和 `familySessionSecret`

### 3.3 必需URL参数

所有请求需要添加以下参数：

```javascript
{
  "clientType": "FAMILY_TV",
  "version": "6.5.5",
  "channelId": "home02",
  "clientSn": "unknown",
  "model": "PJX110",
  "osFamily": "Android",
  "osVersion": "35",
  "networkAccessMode": "WIFI",
  "telecomsOperator": "46011"
}
```

---

## 四、个人云API

### 4.1 文件列表

```http
GET https://api.cloud.189.cn/listFiles.action
```

**必需请求头：** SessionKey签名

**请求参数：**
```
folderId: 文件夹ID（根目录：-11）
fileType: 0（全部类型）
mediaAttr: 0
iconOption: 5
pageNum: 页码（从1开始）
pageSize: 每页数量（最大130）
recursive: 0
orderBy: filename | filesize | lastOpTime
descending: true | false
```

**响应示例：**
```json
{
  "res_code": 0,
  "res_message": "",
  "fileListAO": {
    "count": 10,
    "folderList": [
      {
        "id": "12345",
        "parentId": -11,
        "name": "文件夹名称",
        "lastOpTime": "2024-01-01 12:00:00",
        "createDate": "2024-01-01 12:00:00"
      }
    ],
    "fileList": [
      {
        "id": "67890",
        "name": "文件名.txt",
        "size": 1024,
        "md5": "文件MD5",
        "lastOpTime": "2024-01-01 12:00:00",
        "createDate": "2024-01-01 12:00:00",
        "icon": {
          "smallUrl": "缩略图URL",
          "largeUrl": "大缩略图URL"
        }
      }
    ]
  }
}
```

### 4.2 创建文件夹

```http
POST https://api.cloud.189.cn/createFolder.action
```

**请求参数：**
```
parentFolderId: 父文件夹ID
folderName: 文件夹名称
relativePath: （可选）
```

**响应：**
```json
{
  "id": "新文件夹ID",
  "parentId": 父文件夹ID,
  "name": "文件夹名称",
  "lastOpTime": "2024-01-01 12:00:00",
  "createDate": "2024-01-01 12:00:00"
}
```

### 4.3 获取下载链接

```http
GET https://api.cloud.189.cn/getFileDownloadUrl.action
```

**请求参数：**
```
fileId: 文件ID
dt: 3
flag: 1
```

**响应：**
```json
{
  "fileDownloadUrl": "//download.cloud.189.cn/download/..."
}
```

**获取真实链接：**
```javascript
// 1. 处理URL
url = "https:" + fileDownloadUrl.replace(/&amp;/g, "&")

// 2. GET请求，不允许自动重定向
response = http.get(url, { redirect: false })

// 3. 提取Location头
if (response.status == 302) {
  realUrl = response.headers["Location"]
}
```

### 4.4 重命名文件

```http
POST https://api.cloud.189.cn/renameFile.action
```

**请求参数：**
```
fileId: 文件ID
destFileName: 新文件名
```

### 4.5 重命名文件夹

```http
POST https://api.cloud.189.cn/renameFolder.action
```

**请求参数：**
```
folderId: 文件夹ID
destFolderName: 新文件夹名
```

### 4.6 批量任务操作

#### 创建批量任务

```http
POST https://api.cloud.189.cn/batch/createBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "type": "MOVE | COPY | DELETE",
  "targetFolderId": "目标文件夹ID（DELETE时为空）",
  "taskInfos": "[{\"fileId\":\"ID\",\"fileName\":\"名称\",\"isFolder\":0或1}]"
}
```

**taskInfos格式：**
```json
[
  {
    "fileId": "文件或文件夹ID",
    "fileName": "文件或文件夹名称",
    "isFolder": 0,  // 0=文件，1=文件夹
    "srcParentId": "源父文件夹ID（可选）",
    "dealWay": 0,    // 冲突处理：1=跳过，2=保留，3=覆盖
    "isConflict": 0
  }
]
```

**响应：**
```json
{
  "taskId": "任务ID"
}
```

#### 检查任务状态

```http
POST https://api.cloud.189.cn/batch/checkBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```
type: MOVE | COPY | DELETE
taskId: 任务ID
```

**响应：**
```json
{
  "failedCount": 0,
  "process": 100,
  "skipCount": 0,
  "subTaskCount": 1,
  "successedCount": 1,
  "successedFileIdList": [12345],
  "taskId": "任务ID",
  "taskStatus": 4
}
```

**taskStatus说明：**
- 1: 初始化
- 2: 存在冲突
- 3: 执行中
- 4: 完成

#### 处理冲突

```http
POST https://api.cloud.189.cn/batch/manageBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "targetFolderId": "目标文件夹ID",
  "type": "MOVE | COPY",
  "taskId": "任务ID",
  "taskInfos": "[{\"fileId\":\"ID\",\"fileName\":\"名称\",\"isFolder\":0,\"dealWay\":3}]"
}
```

### 4.7 获取容量信息

```http
GET https://api.cloud.189.cn/portal/getUserSizeInfo.action
```

**响应：**
```json
{
  "res_code": 0,
  "res_message": "",
  "account": "账号",
  "cloudCapacityInfo": {
    "freeSize": 剩余空间,
    "mail189UsedSize": 邮箱使用,
    "totalSize": 总空间,
    "usedSize": 已使用
  },
  "familyCapacityInfo": {
    "freeSize": 家庭云剩余,
    "totalSize": 家庭云总空间,
    "usedSize": 家庭云已使用
  }
}
```

---

## 五、家庭云API

### 5.1 获取家庭云列表

```http
GET https://api.cloud.189.cn/family/manage/getFamilyList.action
```

**签名：** 使用familySessionKey和familySessionSecret

**响应：**
```json
{
  "familyInfoResp": [
    {
      "count": 0,
      "createTime": "创建时间",
      "familyId": 12345,
      "remarkName": "备注名",
      "type": 0,
      "useFlag": 1,
      "userRole": 1
    }
  ]
}
```

**自动选择家庭云：**
```javascript
function getFamilyID(familyInfoResp, loginName) {
  // 优先匹配登录名
  for (info of familyInfoResp) {
    if (loginName.includes(info.remarkName)) {
      return info.familyId.toString()
    }
  }
  // 否则返回第一个
  return familyInfoResp[0].familyId.toString()
}
```

### 5.2 文件列表

```http
GET https://api.cloud.189.cn/family/file/listFiles.action
```

**请求参数：**
```
familyId: 家庭云ID
folderId: 文件夹ID（根目录可能为空）
fileType: 0
mediaAttr: 0
iconOption: 5
pageNum: 页码
pageSize: 每页数量（最大130）
orderBy: 1(文件名) | 2(大小) | 3(时间)
descending: true | false
```

**签名：** 使用familySessionKey和familySessionSecret

### 5.3 创建文件夹

```http
POST https://api.cloud.189.cn/family/file/createFolder.action
```

**请求参数：**
```
familyId: 家庭云ID
parentId: 父文件夹ID
folderName: 文件夹名称
relativePath: （可选）
```

### 5.4 获取下载链接

```http
GET https://api.cloud.189.cn/family/file/getFileDownloadUrl.action
```

**请求参数：**
```
familyId: 家庭云ID
fileId: 文件ID
```

### 5.5 重命名文件

```http
GET https://api.cloud.189.cn/family/file/renameFile.action
```

**请求参数：**
```
familyId: 家庭云ID
fileId: 文件ID
destFileName: 新文件名
```

**注意：** TV端家庭云重命名使用GET方法

### 5.6 重命名文件夹

```http
GET https://api.cloud.189.cn/family/file/renameFolder.action
```

**请求参数：**
```
familyId: 家庭云ID
folderId: 文件夹ID
destFolderName: 新文件夹名
```

### 5.7 批量任务操作

同个人云，但需额外添加 `familyId` 参数。

```http
POST https://api.cloud.189.cn/batch/createBatchTask.action
```

**表单参数：**
```json
{
  "type": "MOVE | COPY | DELETE",
  "familyId": "家庭云ID",
  "targetFolderId": "目标文件夹ID",
  "taskInfos": "[...]"
}
```

### 5.8 清空回收站

```http
POST https://api.cloud.189.cn/batch/createBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "type": "CLEAR_RECYCLE",
  "familyId": "家庭云ID",
  "targetFolderId": "",
  "taskInfos": "[{\"fileId\":\"ID\",\"fileName\":\"名称\",\"isFolder\":0或1}]"
}
```

---

## 六、上传API

### 6.1 上传方式

TV端仅支持旧版上传，不支持新版上传（stream/rapid）。

### 6.2 旧版上传

#### 步骤1: 计算文件MD5

```javascript
async function calculateFileMd5(file) {
  const hash = crypto.createHash('md5')
  const buffer = await file.arrayBuffer()
  hash.update(buffer)
  return hash.digest('hex').toUpperCase()
}
```

#### 步骤2: 创建上传会话

**个人云：**
```http
POST https://api.cloud.189.cn/createUploadFile.action
Content-Type: application/x-www-form-urlencoded
```

**家庭云：**
```http
POST https://api.cloud.189.cn/family/file/createFamilyFile.action
```

**个人云表单参数：**
```json
{
  "parentFolderId": "父文件夹ID",
  "fileName": "文件名",
  "size": "文件大小",
  "md5": "文件MD5",
  "opertype": "3",
  "flag": "1",
  "resumePolicy": "1",
  "isLog": "0"
}
```

**家庭云请求参数：**
```
familyId: 家庭云ID
parentId: 父文件夹ID
fileMd5: 文件MD5
fileName: 文件名
fileSize: 文件大小
resumePolicy: 1
```

**响应：**
```json
{
  "uploadFileId": 上传文件ID,
  "fileUploadUrl": "上传URL",
  "fileCommitUrl": "提交URL",
  "fileDataExists": 0
}
```

**fileDataExists说明：**
- 0: 文件不存在，需要上传
- 1: 文件已存在，秒传成功

#### 步骤3: 检查是否秒传

```javascript
if (fileDataExists == 1) {
  // 秒传成功，直接提交
  return await commitUpload(fileCommitUrl, uploadFileId)
}
```

#### 步骤4: 上传文件

```http
PUT {fileUploadUrl}
```

**必需URL参数：**
```
clientType: FAMILY_TV
version: 6.5.5
channelId: home02
...（其他客户端参数）
```

**请求头：**
```
ResumePolicy: 1
Expect: 100-continue
FamilyId: {家庭云ID}  （家庭云上传时）
UploadFileId: {uploadFileId}  （家庭云上传时）
Edrive-UploadFileId: {uploadFileId}  （个人云上传时）
```

**SessionKey签名：** 需要添加签名请求头
```
Date: {HTTP时间格式}
SessionKey: {sessionKey}
X-Request-ID: {uuid}
Signature: {签名}
```

**请求体：** 文件二进制数据

#### 步骤5: 获取上传状态（断点续传）

**个人云：**
```http
GET https://api.cloud.189.cn/getUploadFileStatus.action
```

**家庭云：**
```http
GET https://api.cloud.189.cn/family/file/getFamilyFileStatus.action
```

**请求参数：**
```
uploadFileId: 上传文件ID
resumePolicy: 1
familyId: {家庭云ID}  （家庭云时）
```

**响应：**
```json
{
  "uploadFileId": 上传文件ID,
  "fileUploadUrl": "上传URL",
  "fileCommitUrl": "提交URL",
  "fileDataExists": 0,
  "dataSize": 已上传大小,
  "size": 当前上传大小
}
```

**断点续传逻辑：**
```javascript
async function uploadWithResume(fileUploadUrl, file, uploadFileId, isFamily) {
  let uploadedSize = 0
  
  while (uploadedSize < file.size) {
    // 上传
    const headers = {
      'ResumePolicy': '1',
      'Expect': '100-continue'
    }
    
    if (isFamily) {
      headers['FamilyId'] = familyId
      headers['UploadFileId'] = uploadFileId.toString()
    } else {
      headers['Edrive-UploadFileId'] = uploadFileId.toString()
    }
    
    await put(fileUploadUrl, file.slice(uploadedSize), headers)
    
    // 检查状态
    const status = await getUploadFileStatus(uploadFileId, isFamily)
    uploadedSize = status.dataSize + status.size
    
    // 秒传成功
    if (status.fileDataExists == 1) {
      break
    }
    
    // 调整文件指针
    file.seek(uploadedSize)
  }
}
```

#### 步骤6: 提交文件

```http
POST {fileCommitUrl}
Content-Type: application/x-www-form-urlencoded
```

**家庭云请求头：**
```
ResumePolicy: 1
UploadFileId: {uploadFileId}
FamilyId: {家庭云ID}
```

**个人云表单参数：**
```json
{
  "opertype": "3",  // 3=覆盖，1=不覆盖
  "resumePolicy": "1",
  "uploadFileId": "上传文件ID",
  "isLog": "0"
}
```

**响应（XML格式）：**
```xml
<file>
  <id>文件ID</id>
  <name>文件名</name>
  <size>文件大小</size>
  <md5>文件MD5</md5>
  <createDate>创建时间</createDate>
</file>
```

### 6.3 秒传

```javascript
async function rapidUpload(parentId, fileMd5, fileName, fileSize, isFamily) {
  // 创建上传会话
  const uploadInfo = await createUploadFile(parentId, fileMd5, fileName, fileSize, isFamily)
  
  // 检查是否秒传
  if (uploadInfo.fileDataExists == 1) {
    // 秒传成功，直接提交
    return await commitUpload(uploadInfo.fileCommitUrl, uploadInfo.uploadFileId, isFamily)
  }
  
  // 秒传失败
  throw new Error('rapid upload fail')
}
```

---

## 七、错误处理

### 7.1 错误响应格式

#### 格式1
```json
{
  "res_code": -1,
  "res_message": "错误信息"
}
```

#### 格式2
```json
{
  "errorCode": "错误码",
  "errorMsg": "错误信息"
}
```

#### 格式3
```json
{
  "code": "ERROR_CODE",
  "message": "错误信息",
  "msg": "错误信息"
}
```

#### 格式4（XML）
```xml
<error>
  <code>错误码</code>
  <message>错误信息</message>
</error>
```

### 7.2 Session失效处理

```javascript
if (response.includes('InvalidSessionKey') || 
    response.includes('userSessionBO is null')) {
  // 尝试刷新Session
  const success = await refreshSession()
  
  if (!success) {
    // Session已过期，需要重新扫码登录
    accessToken = ""
    throw new Error('session expired')
  }
  
  // 重试原请求
  return retry()
}
```

### 7.3 常见错误码

| 错误码/错误信息 | 说明 | 处理方式 |
|----------------|------|---------|
| InvalidSessionKey | Session失效 | 刷新Session或重新登录 |
| userSessionBO is null | Session为空 | 刷新Session |
| FileNotFound | 文件不存在 | 检查文件ID |
| PermissionDenied | 权限不足 | 检查访问权限 |
| QuotaExceeded | 空间不足 | 清理空间 |
| FileAlreadyExists | 文件已存在 | 重命名或覆盖 |

---

## 八、数据结构

### 8.1 文件对象

```typescript
interface Cloud189File {
  id: string;              // 文件ID（字符串）
  name: string;            // 文件名
  size: number;            // 文件大小（字节）
  md5: string;             // 文件MD5
  lastOpTime: string;      // 最后修改时间
  createDate: string;      // 创建时间
  icon: {
    smallUrl: string;      // 小缩略图
    largeUrl: string;      // 大缩略图
    max600: string;        // 最大600px
    mediumUrl: string;     // 中等尺寸
  };
}
```

### 8.2 文件夹对象

```typescript
interface Cloud189Folder {
  id: string;              // 文件夹ID
  parentId: number;        // 父文件夹ID
  name: string;            // 文件夹名
  lastOpTime: string;      // 最后修改时间
  createDate: string;      // 创建时间
}
```

### 8.3 Session信息

```typescript
interface AppSessionResp {
  res_code: number;
  res_message: string;
  loginName: string;
  keepAlive: number;
  getFileDiffSpan: number;
  getUserInfoSpan: number;
  
  // 个人云
  sessionKey: string;
  sessionSecret: string;
  
  // 家庭云
  familySessionKey: string;
  familySessionSecret: string;
  
  // Token
  accessToken: string;
  refreshToken: string;
  
  isSaveName: string;
}
```

### 8.4 家庭云信息

```typescript
interface FamilyInfoResp {
  count: number;
  createTime: string;
  familyId: number;
  remarkName: string;
  type: number;
  useFlag: number;
  userRole: number;
}
```

---

## 九、使用限制

### 9.1 分页限制

- 最大pageSize: 130
- pageNum从1开始

### 9.2 上传限制

- 仅支持旧版上传
- 不支持新版上传（stream/rapid）
- 家庭云不支持覆盖上传

### 9.3 并发限制

- 批量任务间隔:
  - DELETE: 200ms
  - MOVE/COPY: 400ms

### 9.4 Token有效期

- accessToken: 会话级
- 二维码: 约5分钟过期

---

## 十、最佳实践

### 10.1 AccessToken持久化

```javascript
// 登录成功后保存
saveAccessToken(accessToken)

// 初始化时恢复
const savedToken = loadAccessToken()
if (savedToken) {
  try {
    await loginWithAccessToken(savedToken)
  } catch (error) {
    // Token失效，重新扫码
    await qrCodeLogin()
  }
} else {
  await qrCodeLogin()
}
```

### 10.2 Session保活

```javascript
// 启动定时器
setInterval(async () => {
  try {
    await keepAlive()
  } catch (error) {
    await refreshSession()
  }
}, 5 * 60 * 1000)  // 5分钟
```

### 10.3 断点续传上传

```javascript
async function uploadFile(file, parentId, isFamily) {
  // 计算MD5
  const fileMd5 = await calculateFileMd5(file)
  
  // 创建上传会话
  const uploadInfo = await createUploadFile(parentId, fileMd5, file.name, file.size, isFamily)
  
  // 秒传成功
  if (uploadInfo.fileDataExists == 1) {
    return await commitUpload(uploadInfo.fileCommitUrl, uploadInfo.uploadFileId, isFamily)
  }
  
  // 上传文件（支持断点续传）
  await uploadWithResume(uploadInfo.fileUploadUrl, file, uploadInfo.uploadFileId, isFamily)
  
  // 提交
  return await commitUpload(uploadInfo.fileCommitUrl, uploadInfo.uploadFileId, isFamily)
}
```

### 10.4 批量任务等待

```javascript
async function waitBatchTask(type, taskId, interval = 400) {
  while (true) {
    const state = await checkBatchTask(type, taskId)
    
    if (state.taskStatus == 4) {
      return state
    }
    
    if (state.taskStatus == 2) {
      throw new Error('任务冲突')
    }
    
    await sleep(interval)
  }
}
```

---

## 十一、注意事项

1. **仅二维码登录**：TV端不支持密码登录，必须扫码
2. **二维码过期**：二维码约5分钟过期，需要重新获取
3. **签名算法**：登录用AppKey签名，文件操作用SessionKey签名
4. **客户端参数**：所有请求必须携带完整的客户端参数
5. **家庭云签名**：家庭云操作使用familySessionKey签名
6. **覆盖上传**：家庭云不支持覆盖上传
7. **上传方式**：仅支持旧版上传，不支持新版
8. **Session保活**：需要定期调用keepAlive接口

---

## 十二、参考实现

完整的Go语言实现参考：
- 仓库地址：https://github.com/OpenListTeam/OpenList
- 驱动目录：drivers/189_tv/
- 主要文件：
  - driver.go：核心驱动实现
  - utils.go：登录、上传、请求实现
  - help.go：签名算法
  - types.go：数据结构定义