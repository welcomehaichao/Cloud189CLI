# 天翼云盘 PC客户端 API 文档

## 一、概述

PC客户端API是天翼云盘功能最完整的访问方式，支持密码登录、二维码登录、Token持久化、家庭云等完整功能。

### 基本信息

| 项目 | 值 |
|------|-----|
| 网盘主站 | https://cloud.189.cn |
| 认证服务 | https://open.e.189.cn |
| API基础URL | https://api.cloud.189.cn |
| 上传服务 | https://upload.cloud.189.cn |
| 默认根目录ID | -11 |
| AppKey | 8025431004 |
| AccountType | 02 |
| ClientType | 10020 |
| Version | 6.2 |

### 特点

- ✅ 支持密码登录和二维码登录
- ✅ Token持久化，支持自动刷新
- ✅ 支持个人云和家庭云
- ✅ 支持新版上传（stream/rapid）
- ✅ 支持旧版上传
- ✅ 支持秒传
- ✅ 支持家庭云转存个人云

---

## 二、认证机制

### 2.1 密码登录流程

#### 步骤1: 获取登录参数

```http
GET https://cloud.189.cn/api/portal/unifyLoginForPC.action
```

**请求参数：**
```
appId: 8025431004
clientType: 10020
returnURL: https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html
timeStamp: {毫秒时间戳}
```

**响应解析：**
使用正则表达式从HTML中提取：
```javascript
captchaToken = /'captchaToken' value='(.+?)'/.exec(html)[1]
lt = /lt = "(.+?)"/.exec(html)[1]
paramId = /paramId = "(.+?)"/.exec(html)[1]
reqId = /reqId = "(.+?)"/.exec(html)[1]
```

#### 步骤2: 获取RSA加密配置

```http
POST https://open.e.189.cn/api/logbox/config/encryptConf.do
Content-Type: application/x-www-form-urlencoded
Accept: application/json;charset=UTF-8
```

**表单参数：**
```
appId: 8025431004
```

**响应示例：**
```json
{
  "result": 0,
  "data": {
    "upSmsOn": "",
    "pre": "{前缀}",
    "preDomain": "",
    "pubKey": "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC..."
  }
}
```

#### 步骤3: RSA加密用户名密码

```javascript
// 构建PEM格式公钥
publicKey = "-----BEGIN PUBLIC KEY-----\n" + pubKey + "\n-----END PUBLIC KEY-----"

// RSA加密
encryptedUsername = pre + RsaEncrypt(publicKey, username)
encryptedPassword = pre + RsaEncrypt(publicKey, password)

// RsaEncrypt算法
function RsaEncrypt(publicKey, data) {
  // 1. 解析公钥
  pubKey = ParsePKIXPublicKey(publicKey)
  
  // 2. PKCS1v15加密
  encrypted = EncryptPKCS1v15(pubKey, data)
  
  // 3. Hex编码（大写）
  return HexEncode(encrypted).toUpperCase()
}
```

#### 步骤4: 检查是否需要验证码

```http
POST https://open.e.189.cn/api/logbox/oauth2/needcaptcha.do
Content-Type: application/x-www-form-urlencoded
```

**请求头：**
```
REQID: {reqId}
```

**表单参数：**
```
appKey: 8025431004
accountType: 02
userName: {加密后的用户名}
```

**响应：**
- "0": 不需要验证码
- 其他: 需要验证码

#### 步骤5: 获取验证码（如需要）

```http
GET https://open.e.189.cn/api/logbox/oauth2/picCaptcha.do
```

**请求参数：**
```
token: {captchaToken}
REQID: {reqId}
rnd: {毫秒时间戳}
```

**响应：** PNG图片二进制数据

#### 步骤6: 提交登录

```http
POST https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do
Content-Type: application/x-www-form-urlencoded
Accept: application/json;charset=UTF-8
Force-Content-Type: application/json;charset=UTF-8
```

**请求头：**
```
REQID: {reqId}
lt: {lt}
```

**表单参数：**
```json
{
  "appKey": "8025431004",
  "accountType": "02",
  "userName": "{pre}{RSA加密的用户名}",
  "password": "{pre}{RSA加密的密码}",
  "validateCode": "{验证码或空}",
  "captchaToken": "{captchaToken}",
  "returnUrl": "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html",
  "dynamicCheck": "FALSE",
  "clientType": "10020",
  "cb_SaveName": "1",
  "isOauth2": "false",
  "state": "",
  "paramId": "{paramId}"
}
```

**响应示例：**
```json
{
  "result": 0,
  "msg": "登录成功",
  "toUrl": "https://cloud.189.cn/..."
}
```

#### 步骤7: 获取Session

```http
POST https://api.cloud.189.cn/getSessionForPC.action
Accept: application/json;charset=UTF-8
```

**请求参数：**
```
redirectURL: {toUrl}
clientType: TELEPC
version: 6.2
channelId: web_cloud.189.cn
rand: {random1}_{random2}
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
- `refreshToken`: 刷新令牌（可持久化保存）

---

### 2.2 二维码登录流程

#### 步骤1: 获取基础参数

同密码登录步骤1

#### 步骤2: 获取二维码UUID

```http
POST https://open.e.189.cn/api/logbox/oauth2/getUUID.do
Content-Type: application/x-www-form-urlencoded
Accept: application/json;charset=UTF-8
Force-Content-Type: application/json;charset=UTF-8
```

**表单参数：**
```
appId: 8025431004
```

**响应示例：**
```json
{
  "uuid": "https://m.cloud.189.cn/zhuanti/qrLogin/qrCode/index.html?uuid={uuid}",
  "encodeuuid": "{encodeuuid}",
  "encryuuid": "{encryuuid}"
}
```

#### 步骤3: 生成并展示二维码

```javascript
// 生成二维码图片
qrCode = QRCode(uuid, 256x256)

// 展示给用户
displayQRCode(qrCode)

// 提供扫码链接
link = uuid
```

#### 步骤4: 检查扫码状态

```http
POST https://open.e.189.cn/api/logbox/oauth2/qrcodeLoginState.do
Content-Type: application/x-www-form-urlencoded
Accept: application/json;charset=UTF-8
```

**请求头：**
```
Referer: https://open.e.189.cn
Reqid: {reqId}
lt: {lt}
```

**表单参数：**
```json
{
  "appId": "8025431004",
  "clientType": "10020",
  "returnUrl": "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html",
  "paramId": "{paramId}",
  "uuid": "{uuid}",
  "encryuuid": "{encryuuid}",
  "date": "20240101120000.000",
  "timeStamp": "{毫秒时间戳}"
}
```

**响应示例：**
```json
{
  "status": 0,        // 状态码
  "redirectUrl": "",  // 登录成功后的跳转URL
  "msg": ""
}
```

**状态码说明：**
| status | 说明 | 处理 |
|--------|------|------|
| 0 | 登录成功 | 继续获取Session |
| -106 | 等待扫码 | 继续轮询 |
| -11002 | 已扫码，等待确认 | 继续轮询 |
| -11001 | 二维码过期 | 重新获取UUID |

#### 步骤5: 获取Session

同密码登录步骤7

---

### 2.3 Token刷新机制

#### 使用refreshToken刷新

```http
POST https://open.e.189.cn/api/oauth2/refreshToken.do
Content-Type: application/x-www-form-urlencoded
Accept: application/json;charset=UTF-8
Force-Content-Type: application/json;charset=UTF-8
```

**表单参数：**
```json
{
  "clientId": "8025431004",
  "refreshToken": "{refreshToken}",
  "grantType": "refresh_token",
  "format": "json"
}
```

**响应示例：**
```json
{
  "res_code": 0,
  "sessionKey": "{sessionKey}",
  "sessionSecret": "{sessionSecret}",
  "familySessionKey": "{familySessionKey}",
  "familySessionSecret": "{familySessionSecret}",
  "accessToken": "{newAccessToken}",
  "refreshToken": "{newRefreshToken}"
}
```

**注意：** 
- 刷新成功后需要更新保存的refreshToken
- 刷新失败需要重新登录

#### Session保活

```http
GET https://api.cloud.189.cn/keepUserSession.action
```

**请求参数：**
```
clientType: TELEPC
version: 6.2
channelId: web_cloud.189.cn
rand: {random}
```

**建议频率：** 每5分钟执行一次

#### Session刷新

```http
GET https://api.cloud.189.cn/getSessionForPC.action
```

**请求参数：**
```
appId: 8025431004
accessToken: {accessToken}
clientType: TELEPC
version: 6.2
channelId: web_cloud.189.cn
rand: {random}
```

**请求头：**
```
X-Request-ID: {uuid}
```

---

## 三、签名机制

### 3.1 SessionKey签名

用于所有需要认证的API请求。

#### 签名算法

```javascript
function Signature(sessionSecret, sessionKey, method, url, date, params) {
  // 1. 提取URI路径
  urlpath = url.match(/:\/\/[^\/]+((\/[^\/\s?#]+)*)/)[1]
  
  // 2. 构建签名字符串
  data = "SessionKey=" + sessionKey
  data += "&Operate=" + method
  data += "&RequestURI=" + urlpath
  data += "&Date=" + date
  
  // 3. 如果有加密参数，添加params
  if (params && params != "") {
    data += "&params=" + params
  }
  
  // 4. HMAC-SHA1签名
  signature = HMAC_SHA1(sessionSecret, data)
  
  // 5. Hex编码（大写）
  return HexEncode(signature).toUpperCase()
}
```

#### 签名示例

```javascript
sessionSecret = "abc123..."
sessionKey = "xyz789..."
method = "GET"
url = "https://api.cloud.189.cn/listFiles.action"
date = "Mon, 01 Jan 2024 12:00:00 GMT"
params = "加密后的参数或空"

signature = Signature(sessionSecret, sessionKey, method, url, date, params)
// 结果: "A1B2C3D4E5F6..."
```

### 3.2 AES参数加密

部分接口需要对请求参数进行AES加密。

#### 加密算法

```javascript
function AesEncrypt(params, sessionSecret) {
  // 1. 取sessionSecret前16位作为密钥
  key = sessionSecret.substring(0, 16)
  
  // 2. 参数排序并拼接
  queryString = ParamsEncode(params)
  // 格式: key1=value1&key2=value2...（按key字典序排列）
  
  // 3. PKCS7填充
  data = PKCS7Padding(queryString, 16)
  
  // 4. AES-ECB加密
  encrypted = AES_ECB_Encrypt(data, key)
  
  // 5. Hex编码（大写）
  return HexEncode(encrypted).toUpperCase()
}
```

#### 参数编码规则

```javascript
function ParamsEncode(params) {
  // 1. 按key字典序排序
  keys = Object.keys(params).sort()
  
  // 2. 拼接
  parts = []
  for (key of keys) {
    parts.push(key + "=" + params[key])
  }
  
  return parts.join("&")
}
```

### 3.3 必需请求头

```http
Accept: application/json;charset=UTF-8
Date: {HTTP时间格式}
SessionKey: {sessionKey或familySessionKey}
X-Request-ID: {uuid}
Signature: {签名}
```

#### 时间格式

```
Date: Mon, 01 Jan 2024 12:00:00 GMT
```

Go实现：
```go
date := time.Now().UTC().Format(http.TimeFormat)
```

### 3.4 必需URL参数

所有请求需要添加以下参数：

```javascript
{
  "clientType": "TELEPC",
  "version": "6.2",
  "channelId": "web_cloud.189.cn",
  "rand": "{random1}_{random2}"
}
```

示例：
```
rand=12345_6789012345
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
pageSize: 每页数量（最大1000）
recursive: 0
orderBy: filename | filesize | lastOpTime
descending: true | false
```

**签名说明：**
- Method: GET
- URL: https://api.cloud.189.cn/listFiles.action
- Params: 无需加密

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

**响应：** 返回新的文件对象

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
  "process": 100,           // 进度百分比
  "skipCount": 0,
  "subTaskCount": 1,
  "successedCount": 1,
  "successedFileIdList": [12345],
  "taskId": "任务ID",
  "taskStatus": 4           // 1=初始化，2=冲突，3=执行中，4=完成
}
```

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

#### 等待任务完成

```javascript
async function waitBatchTask(type, taskId, interval = 400) {
  while (true) {
    const state = await checkBatchTask(type, taskId)
    
    switch (state.taskStatus) {
      case 2: // 冲突
        throw new Error('任务冲突')
      case 4: // 完成
        return state
    }
    
    await sleep(interval)
  }
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
pageSize: 每页数量
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

**注意：** 家庭云重命名使用GET方法

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

### 5.8 保存家庭云文件到个人云

```http
POST https://api.cloud.189.cn/batch/createBatchTask.action
```

**表单参数：**
```json
{
  "type": "COPY",
  "familyId": "家庭云ID",
  "targetFolderId": "个人云目标文件夹ID",
  "copyType": "2",
  "groupId": "null",
  "shareId": "null",
  "taskInfos": "[{\"fileId\":\"家庭云文件ID\",\"fileName\":\"名称\",\"isFolder\":0}]"
}
```

**等待转存完成：**
```javascript
async function saveFamilyFileToPersonCloud(familyId, srcObj, dstDir, overwrite) {
  const task = {
    fileId: srcObj.id,
    fileName: srcObj.name,
    isFolder: srcObj.isDir ? 1 : 0
  }
  
  const resp = await createBatchTask("COPY", familyId, dstDir.id, {
    groupId: "null",
    copyType: "2",
    shareId: "null"
  }, task)
  
  while (true) {
    const state = await checkBatchTask("COPY", resp.taskId)
    
    switch (state.taskStatus) {
      case 2: // 冲突
        task.dealWay = overwrite ? 3 : 2
        await manageBatchTask("COPY", resp.taskId, dstDir.id, task)
        break
      case 4: // 完成
        return
    }
    
    await sleep(400)
  }
}
```

---

## 六、上传API

### 6.1 上传方式对比

| 方式 | 说明 | 适用场景 |
|------|------|---------|
| stream | 流式上传，边读边传 | 大文件，内存受限 |
| rapid | 计算完整MD5后上传 | 需要秒传 |
| old | 旧版上传，兼容性好 | 需要兼容旧接口 |

### 6.2 新版上传（stream模式）

#### 步骤1: 计算分片大小

```javascript
function partSize(fileSize) {
  const DEFAULT = 10 * 1024 * 1024  // 10MB
  
  if (fileSize > DEFAULT * 2 * 999) {
    // 大文件动态计算，最大1999片
    return Math.max(
      Math.ceil(fileSize / 1999 / DEFAULT),
      5
    ) * DEFAULT
  }
  
  if (fileSize > DEFAULT * 999) {
    return DEFAULT * 2  // 20MB
  }
  
  return DEFAULT  // 10MB
}
```

#### 步骤2: 初始化上传

```http
GET https://upload.cloud.189.cn/person/initMultiUpload
```

**AES加密参数：**
```json
{
  "parentFolderId": "父文件夹ID",
  "fileName": "URL编码的文件名",
  "fileSize": "文件大小",
  "sliceSize": "分片大小",
  "lazyCheck": "1"
}
```

**响应：**
```json
{
  "code": "SUCCESS",
  "data": {
    "uploadType": 1,
    "uploadHost": "上传服务器",
    "uploadFileId": "上传文件ID",
    "fileDataExists": 0
  }
}
```

#### 步骤3: 计算分片MD5

```javascript
async function calculatePartMd5(file, sliceSize) {
  const count = Math.ceil(file.size / sliceSize)
  const md5s = []
  const fileMd5 = crypto.createHash('md5')
  
  for (let i = 0; i < count; i++) {
    const start = i * sliceSize
    const end = Math.min(start + sliceSize, file.size)
    const slice = file.slice(start, end)
    const buffer = await slice.arrayBuffer()
    
    const sliceMd5 = crypto.createHash('md5')
    sliceMd5.update(buffer)
    md5s.push(sliceMd5.digest('hex').toUpperCase())
    fileMd5.update(buffer)
  }
  
  return {
    fileMd5: fileMd5.digest('hex').toUpperCase(),
    sliceMd5s: md5s
  }
}
```

#### 步骤4: 获取上传URL

```http
GET https://upload.cloud.189.cn/person/getMultiUploadUrls
```

**AES加密参数：**
```json
{
  "uploadFileId": "上传文件ID",
  "partInfo": "分片编号-MD5的Base64编码"
}
```

**partInfo格式：**
```
1-AbCdEf123456==
```

多个分片用逗号分隔：
```
1-AbCdEf123456==,2-GhIjKl789012==
```

**响应：**
```json
{
  "code": "SUCCESS",
  "uploadUrls": {
    "partNumber_1": {
      "requestURL": "https://...",
      "requestHeader": "Authorization=xxx&Content-Type=application/octet-stream"
    },
    "partNumber_2": {
      "requestURL": "https://...",
      "requestHeader": "..."
    }
  }
}
```

#### 步骤5: 上传分片

```http
PUT {requestURL}
```

**请求头解析：**
```javascript
function parseHeaders(requestHeader) {
  const headers = {}
  const parts = requestHeader.split('&')
  for (const part of parts) {
    const [key, value] = part.split('=')
    headers[key] = value
  }
  return headers
}
```

**必需URL参数：**
```
clientType: TELEPC
version: 6.2
channelId: web_cloud.189.cn
rand: {random}
```

**必需请求头：**
```
Authorization: {从requestHeader解析}
Content-Type: application/octet-stream
```

**请求体：** 分片二进制数据

#### 步骤6: 提交上传

```http
GET https://upload.cloud.189.cn/person/commitMultiUploadFile
```

**AES加密参数：**
```json
{
  "uploadFileId": "上传文件ID",
  "fileMd5": "文件完整MD5",
  "sliceMd5": "分片MD5组合后的MD5或fileMd5",
  "lazyCheck": "1",
  "isLog": "0",
  "opertype": "3"  // 3=覆盖，1=不覆盖
}
```

**sliceMd5计算：**
```javascript
function calculateSliceMd5(fileMd5, sliceMd5s, fileSize, sliceSize) {
  if (fileSize <= sliceSize) {
    return fileMd5
  }
  
  // 所有分片MD5用换行符连接，再计算MD5
  const combined = sliceMd5s.join('\n')
  return crypto.createHash('md5').update(combined).digest('hex').toUpperCase()
}
```

**响应：**
```json
{
  "file": {
    "userFileId": "文件ID",
    "fileName": "文件名",
    "fileSize": 文件大小,
    "fileMd5": "文件MD5",
    "createDate": "创建时间"
  }
}
```

### 6.3 快速上传（rapid模式）

适用于需要秒传的场景。

#### 步骤1: 计算文件完整MD5

```javascript
const fileMd5 = await calculateFileMd5(file)
```

#### 步骤2: 初始化上传

```http
GET https://upload.cloud.189.cn/person/initMultiUpload
```

**AES加密参数：**
```json
{
  "parentFolderId": "父文件夹ID",
  "fileName": "文件名",
  "fileSize": "文件大小",
  "fileMd5": "文件完整MD5",
  "sliceSize": "分片大小",
  "sliceMd5": "分片MD5"
}
```

**响应：**
```json
{
  "code": "SUCCESS",
  "data": {
    "uploadFileId": "上传文件ID",
    "fileDataExists": 1  // 1表示秒传成功
  }
}
```

#### 步骤3: 检查是否秒传

```javascript
if (fileDataExists == 1) {
  // 秒传成功，直接提交
  return await commitUpload(uploadFileId, overwrite)
} else {
  // 需要实际上传
  return await uploadParts(...)
}
```

### 6.4 旧版上传

#### 步骤1: 创建上传会话

```http
POST https://api.cloud.189.cn/createUploadFile.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
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

**响应：**
```json
{
  "uploadFileId": 上传文件ID,
  "fileUploadUrl": "上传URL",
  "fileCommitUrl": "提交URL",
  "fileDataExists": 0
}
```

#### 步骤2: 上传文件

```http
PUT {fileUploadUrl}
```

**请求头：**
```
ResumePolicy: 1
Expect: 100-continue
Edrive-UploadFileId: {uploadFileId}
```

**签名：** 需要SessionKey签名

**请求体：** 文件二进制数据

#### 步骤3: 获取上传状态（断点续传）

```http
GET https://api.cloud.189.cn/getUploadFileStatus.action
```

**请求参数：**
```
uploadFileId: 上传文件ID
resumePolicy: 1
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

**断点续传：**
```javascript
async function uploadWithResume(fileUploadUrl, file, uploadFileId) {
  let uploadedSize = 0
  
  while (uploadedSize < file.size) {
    // 上传
    await put(fileUploadUrl, file.slice(uploadedSize), {
      'ResumePolicy': '1',
      'Edrive-UploadFileId': uploadFileId.toString()
    })
    
    // 检查状态
    const status = await getUploadFileStatus(uploadFileId)
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

#### 步骤4: 提交文件

```http
POST {fileCommitUrl}
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
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

### 6.5 家庭云上传

家庭云上传URL替换为：
- 初始化：`https://upload.cloud.189.cn/family/initMultiUpload`
- 获取URL：`https://upload.cloud.189.cn/family/getMultiUploadUrls`
- 提交：`https://upload.cloud.189.cn/family/commitMultiUploadFile`
- 旧版创建：`https://api.cloud.189.cn/family/file/createFamilyFile.action`

**额外参数：**
```
familyId: 家庭云ID
```

**注意：** 家庭云不支持覆盖上传

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
    // 需要重新登录
    await login()
  }
  
  // 重试原请求
  return retry()
}
```

### 7.3 常见错误码

| 错误码/错误信息 | 说明 | 处理方式 |
|----------------|------|---------|
| InvalidSessionKey | Session失效 | 刷新Session或重新登录 |
| UserInvalidOpenToken | Token无效 | 使用refreshToken刷新或重新登录 |
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

### 8.4 批量任务信息

```typescript
interface BatchTaskInfo {
  fileId: string;
  fileName: string;
  isFolder: number;        // 0=文件，1=文件夹
  srcParentId?: string;
  dealWay?: number;        // 1=跳过，2=保留，3=覆盖
  isConflict?: number;
}

interface BatchTaskStateResp {
  failedCount: number;
  process: number;         // 进度百分比
  skipCount: number;
  subTaskCount: number;
  successedCount: number;
  successedFileIdList: number[];
  taskId: string;
  taskStatus: number;      // 1=初始化，2=冲突，3=执行中，4=完成
}
```

---

## 九、使用限制

### 9.1 分页限制

- 最大pageSize: 1000
- pageNum从1开始

### 9.2 上传限制

- 默认分片大小: 10MB
- 分片数量限制:
  - 10MB/20MB: 最大999片
  - 50MB以上: 最大1999片
- 上传线程: 建议1-32

### 9.3 并发限制

- 批量任务间隔:
  - DELETE: 200ms
  - MOVE/COPY: 400ms

### 9.4 Token有效期

- accessToken: 会话级
- refreshToken: 可长期保存
- Session保活: 每5分钟

---

## 十、最佳实践

### 10.1 Token持久化

```javascript
// 登录成功后保存
saveToken({
  refreshToken: response.refreshToken,
  sessionKey: response.sessionKey,
  sessionSecret: response.sessionSecret
})

// 初始化时恢复
const savedToken = loadToken()
if (savedToken.refreshToken) {
  await refreshToken(savedToken.refreshToken)
} else {
  await login()
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
// 保存上传进度
saveUploadProgress({
  uploadFileId,
  uploadParts: ['part1-info', 'part2-info']
})

// 恢复上传
const progress = loadUploadProgress()
if (progress) {
  const parts = progress.uploadParts.filter(p => p != '')
  await uploadRemainingParts(parts)
}
```

### 10.4 批量任务重试

```javascript
async function waitBatchTask(type, taskId, interval = 400) {
  while (true) {
    const state = await checkBatchTask(type, taskId)
    
    if (state.taskStatus == 4) {
      return state
    }
    
    if (state.taskStatus == 2) {
      // 冲突处理
      await manageBatchTask(type, taskId, targetFolderId, {
        fileId: srcFile.id,
        fileName: srcFile.name,
        isFolder: 0,
        dealWay: 3  // 覆盖
      })
    }
    
    await sleep(interval)
  }
}
```

---

## 十一、注意事项

1. **签名算法**：所有需要认证的请求必须携带签名
2. **参数加密**：上传接口需要AES加密参数
3. **时间格式**：Date使用HTTP规范时间格式（GMT）
4. **家庭云签名**：家庭云操作使用familySessionKey签名
5. **覆盖上传**：家庭云不支持覆盖上传
6. **二维码过期**：二维码约5分钟过期，需要重新获取
7. **Token刷新**：refreshToken刷新后需要更新保存
8. **批量任务**：频繁创建批量任务可能导致失败

---

## 十二、参考实现

完整的Go语言实现参考：
- 仓库地址：https://github.com/OpenListTeam/OpenList
- 驱动目录：drivers/189pc/
- 主要文件：
  - driver.go：核心驱动实现
  - utils.go：登录、上传、请求实现
  - help.go：签名算法
  - types.go：数据结构定义