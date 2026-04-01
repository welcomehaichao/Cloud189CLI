# 天翼云盘 Web版 API 文档

## 一、概述

Web版API是天翼云盘最基础的访问方式，通过网页端认证机制实现。适用于需要模拟浏览器行为的场景。

### 基本信息

| 项目 | 值 |
|------|-----|
| 网盘主站 | https://cloud.189.cn |
| 认证服务 | https://open.e.189.cn |
| API基础URL | https://cloud.189.cn/api |
| 上传服务 | https://upload.cloud.189.cn |
| 默认根目录ID | -11 |

### 特点

- ✅ 基础文件操作完整
- ✅ 支持新版和旧版上传
- ❌ 无Token持久化机制
- ❌ 不支持家庭云
- ❌ 不支持秒传

---

## 二、认证机制

### 2.1 登录流程

Web版登录采用OAuth2.0流程，需要多个步骤完成认证。

#### 步骤1: 获取登录入口URL

```http
GET https://cloud.189.cn/api/portal/loginUrl.action?redirectURL=https%3A%2F%2Fcloud.189.cn%2Fmain.action
```

**响应说明：**
- 如果已登录，会重定向到 `https://cloud.189.cn/web/main`
- 如果未登录，会重定向到登录页面，URL中包含关键参数

**关键参数获取：**
```
lt: 登录令牌
reqId: 请求ID  
appId: 应用ID（每次动态生成）
```

#### 步骤2: 获取应用配置

```http
POST https://open.e.189.cn/api/logbox/oauth2/appConf.do
Content-Type: application/x-www-form-urlencoded

version=2.0&appKey={appId}
```

**请求头：**
```
lt: {lt}
reqid: {reqId}
referer: {登录页面URL}
origin: https://open.e.189.cn
```

**响应示例：**
```json
{
  "result": "0",
  "msg": "",
  "data": {
    "accountType": "01",
    "appKey": "{appId}",
    "clientType": 10010,
    "isOauth2": false,
    "loginSort": "",
    "mailSuffix": "@pan.cn",
    "pageKey": "",
    "paramId": "{paramId}",
    "regReturnUrl": "",
    "reqId": "{reqId}",
    "returnUrl": "{returnUrl}",
    "showFeedback": "",
    "showPwSaveName": "",
    "showQrSaveName": "",
    "showSmsSaveName": "",
    "sso": ""
  }
}
```

#### 步骤3: 获取加密配置

```http
POST https://open.e.189.cn/api/logbox/config/encryptConf.do
Content-Type: application/x-www-form-urlencoded

appId={appId}
```

**响应示例：**
```json
{
  "result": 0,
  "data": {
    "upSmsOn": "",
    "pre": "{加密前缀}",
    "preDomain": "",
    "pubKey": "{RSA公钥Base64}"
  }
}
```

#### 步骤4: 提交登录

```http
POST https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do
Content-Type: application/x-www-form-urlencoded
```

**请求头：**
```
lt: {lt}
reqid: {reqId}
referer: {登录页面URL}
origin: https://open.e.189.cn
```

**表单参数：**
```json
{
  "version": "v2.0",
  "apToken": "",
  "appKey": "{appId}",
  "accountType": "01",
  "userName": "{pre}{RSA加密的用户名}",
  "epd": "{pre}{RSA加密的密码}",
  "captchaType": "",
  "validateCode": "",
  "smsValidateCode": "",
  "captchaToken": "{captchaToken}",
  "returnUrl": "{returnUrl}",
  "mailSuffix": "@pan.cn",
  "dynamicCheck": "FALSE",
  "clientType": "10010",
  "cb_SaveName": "3",
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
  "toUrl": "https://cloud.189.cn/web/main?..."
}
```

**错误处理：**
- `result != 0`: 登录失败，msg包含错误信息
- 验证码错误需要重新获取验证码图片

#### 步骤5: 完成认证

访问返回的 `toUrl` 完成认证，系统会设置必要的Cookie。

```http
GET {toUrl}
```

### 2.2 RSA加密算法

```javascript
// 加密格式
加密用户名 = pre + RSA_Encrypt_Hex(username, pubKey)
加密密码 = pre + RSA_Encrypt_Hex(password, pubKey)

// RSA加密步骤
1. 解析PEM格式公钥：
   publicKey = "-----BEGIN PUBLIC KEY-----\n" + pubKey + "\n-----END PUBLIC KEY-----"
   
2. 使用PKCS1v15加密：
   encryptedData = RSA_Encrypt_PKCS1v15(publicKey, data)
   
3. Base64编码：
   base64Str = Base64_Encode(encryptedData)
   
4. 转换为Hex（大写）：
   hexStr = Base64ToHex(base64Str)
```

**Base64转Hex算法：**
```javascript
function b64tohex(a) {
  const b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
  const BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz";
  
  let d = "";
  let e = 0;
  let c = 0;
  
  for (let i = 0; i < a.length; i++) {
    const m = a[i];
    if (m != "=") {
      const v = b64map.indexOf(m);
      if (e == 0) {
        e = 1;
        d += BI_RM[v >> 2];
        c = 3 & v;
      } else if (e == 1) {
        e = 2;
        d += BI_RM[c << 2 | v >> 4];
        c = 15 & v;
      } else if (e == 2) {
        e = 3;
        d += BI_RM[c];
        d += BI_RM[v >> 2];
        c = 3 & v;
      } else {
        e = 0;
        d += BI_RM[c << 2 | v >> 4];
        d += BI_RM[15 & v];
      }
    }
  }
  
  if (e == 1) {
    d += BI_RM[c << 2];
  }
  
  return d.toUpperCase();
}
```

### 2.3 验证码处理

#### 获取验证码图片

```http
GET https://open.e.189.cn/api/logbox/oauth2/picCaptcha.do?token={captchaToken}&timeStamp={timestamp}
```

**响应：** PNG图片二进制数据

#### OCR识别或手动输入

1. 使用OCR API识别验证码
2. 或返回给用户手动输入

---

## 三、签名与加密

### 3.1 请求基础参数

所有API请求需要添加随机参数：

```
?noCache=0.{17位随机数}
```

示例：
```
noCache=0.12345678901234567
```

### 3.2 SessionKey签名（上传接口）

上传相关接口需要SessionKey签名。

#### 获取SessionKey

```http
GET https://cloud.189.cn/v2/getUserBriefInfo.action?noCache={random}
```

**响应：**
```json
{
  "sessionKey": "{sessionKey}"
}
```

#### 获取RSA密钥

```http
GET https://cloud.189.cn/api/security/generateRsaKey.action?noCache={random}
```

**响应：**
```json
{
  "pubKey": "{pubKey}",
  "pkId": "{pkId}",
  "expire": {过期时间戳}
}
```

#### AES加密参数

```javascript
// 生成随机密钥
l = RandomString(16-32位)  // 取前16位作为AES密钥

// 加密params
data = AES_ECB_Encrypt(paramsQueryString, l[0:16])
encryptedParams = HexEncode(data)

// HMAC签名
signature = HMAC_SHA1(
  "SessionKey={sessionKey}&Operate={method}&RequestURI={uri}&Date={date}&params={encryptedParams}",
  l
)
```

#### 上传请求头

```
accept: application/json;charset=UTF-8
SessionKey: {sessionKey}
Signature: {signature}
X-Request-Date: {timestamp_ms}
X-Request-ID: {uuid}
EncryptionText: {RSA加密的l密钥}
PkId: {pkId}
```

---

## 四、文件操作API

### 4.1 获取文件列表

```http
GET https://cloud.189.cn/api/open/file/listFiles.action
```

**请求参数：**
```
folderId: 文件夹ID（根目录：-11）
pageSize: 60
pageNum: 页码（从1开始）
mediaType: 0（全部类型）
iconOption: 5（获取缩略图）
orderBy: lastOpTime | filename | filesize
descending: true | false
noCache: {random}
```

**响应示例：**
```json
{
  "res_code": 0,
  "res_message": "success",
  "fileListAO": {
    "count": 10,
    "folderList": [
      {
        "id": 12345,
        "name": "文件夹名称",
        "lastOpTime": "2024-01-01 12:00:00"
      }
    ],
    "fileList": [
      {
        "id": 67890,
        "name": "文件名.txt",
        "size": 1024,
        "lastOpTime": "2024-01-01 12:00:00",
        "icon": {
          "smallUrl": "缩略图URL"
        }
      }
    ]
  }
}
```

### 4.2 创建文件夹

```http
POST https://cloud.189.cn/api/open/file/createFolder.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```
parentFolderId: 父文件夹ID
folderName: 新文件夹名称
```

**响应：**
```json
{
  "res_code": 0,
  "res_message": "success"
}
```

### 4.3 获取下载链接

```http
GET https://cloud.189.cn/api/portal/getFileInfo.action
```

**请求参数：**
```
fileId: 文件ID
noCache: {random}
```

**响应：**
```json
{
  "res_code": 0,
  "res_message": "",
  "downloadUrl": "//download.cloud.189.cn/download/..."
}
```

**获取真实链接：**
```
1. 拼接协议：https:{downloadUrl}
2. GET请求，允许302重定向
3. 第一个302的location即为真实下载链接
4. 可能需要多次302重定向
```

### 4.4 重命名文件

```http
POST https://cloud.189.cn/api/open/file/renameFile.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```
fileId: 文件ID
destFileName: 新文件名
```

### 4.5 重命名文件夹

```http
POST https://cloud.189.cn/api/open/file/renameFolder.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```
folderId: 文件夹ID
destFolderName: 新文件夹名
```

### 4.6 移动文件/文件夹

```http
POST https://cloud.189.cn/api/open/batch/createBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "type": "MOVE",
  "targetFolderId": "目标文件夹ID",
  "taskInfos": "[{\"fileId\":\"源ID\",\"fileName\":\"名称\",\"isFolder\":0或1}]"
}
```

**isFolder说明：**
- 0: 文件
- 1: 文件夹

### 4.7 复制文件/文件夹

```http
POST https://cloud.189.cn/api/open/batch/createBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "type": "COPY",
  "targetFolderId": "目标文件夹ID",
  "taskInfos": "[{\"fileId\":\"源ID\",\"fileName\":\"名称\",\"isFolder\":0或1}]"
}
```

### 4.8 删除文件/文件夹

```http
POST https://cloud.189.cn/api/open/batch/createBatchTask.action
Content-Type: application/x-www-form-urlencoded
```

**表单参数：**
```json
{
  "type": "DELETE",
  "targetFolderId": "",
  "taskInfos": "[{\"fileId\":\"ID\",\"fileName\":\"名称\",\"isFolder\":0或1}]"
}
```

### 4.9 获取容量信息

```http
GET https://cloud.189.cn/api/portal/getUserSizeInfo.action?noCache={random}
```

**响应：**
```json
{
  "res_code": 0,
  "res_message": "",
  "account": "账号",
  "cloudCapacityInfo": {
    "freeSize": 剩余空间,
    "mail189UsedSize": 189邮箱使用,
    "totalSize": 总空间,
    "usedSize": 已使用
  }
}
```

---

## 五、上传API

### 5.1 新版上传（推荐）

#### 初始化上传

```http
GET https://upload.cloud.189.cn/person/initMultiUpload
```

**加密参数（AES加密后）：**
```json
{
  "parentFolderId": "父文件夹ID",
  "fileName": "URL编码的文件名",
  "fileSize": "文件大小",
  "sliceSize": "分片大小（默认10485760）",
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

**fileDataExists说明：**
- 0: 文件不存在，需要上传
- 1: 文件已存在，可以秒传（但Web版不支持）

#### 获取上传URL

```http
GET https://upload.cloud.189.cn/person/getMultiUploadUrls
```

**加密参数：**
```json
{
  "partInfo": "分片编号-分片MD5的Base64编码",
  "uploadFileId": "上传文件ID"
}
```

**示例：**
```
partInfo: 1-AbCdEf123456==
```

**响应：**
```json
{
  "code": "SUCCESS",
  "uploadUrls": {
    "partNumber_1": {
      "requestURL": "上传URL",
      "requestHeader": "Authorization=xxx&Content-Type=application/octet-stream"
    }
  }
}
```

#### 上传分片

```http
PUT {requestURL}
```

**请求头：** 从requestHeader解析
```
Authorization: xxx
Content-Type: application/octet-stream
```

**请求体：** 分片二进制数据

#### 提交上传

```http
GET https://upload.cloud.189.cn/person/commitMultiUploadFile
```

**加密参数：**
```json
{
  "uploadFileId": "上传文件ID",
  "fileMd5": "文件完整MD5",
  "sliceMd5": "分片MD5组合后的MD5",
  "lazyCheck": "1",
  "opertype": "3"  // 3=覆盖，1=不覆盖
}
```

**sliceMd5计算规则：**
```
如果 fileSize <= sliceSize:
  sliceMd5 = fileMd5
  
如果 fileSize > sliceSize:
  sliceMd5 = MD5(所有分片MD5用换行符连接)
  
示例：
  sliceMd5 = MD5("ABCD1234\nEFGH5678\n...")
```

### 5.2 旧版上传

```http
POST https://hb02.upload.cloud.189.cn/v1/DCIWebUploadAction
Content-Type: multipart/form-data
```

**表单参数：**
```
parentId: 父文件夹ID
sessionKey: sessionKey（需获取）
opertype: 1
fname: 文件名
```

**文件字段：**
```
Filedata: 文件二进制数据
```

**响应：**
```json
{
  "MD5": "文件MD5",
  "id": "文件ID"
}
```

---

## 六、错误处理

### 6.1 错误响应格式

```json
{
  "errorCode": "错误码",
  "errorMsg": "错误信息"
}
```

或：

```json
{
  "res_code": -1,
  "res_message": "错误信息"
}
```

### 6.2 常见错误码

| 错误码 | 说明 | 处理方式 |
|-------|------|---------|
| InvalidSessionKey | Session失效 | 重新登录 |
| FileNotFound | 文件不存在 | 检查文件ID |
| PermissionDenied | 权限不足 | 检查访问权限 |
| QuotaExceeded | 空间不足 | 清理空间或扩容 |
| FileAlreadyExists | 文件已存在 | 重命名或覆盖 |

### 6.3 Session失效处理

```javascript
if (errorCode == "InvalidSessionKey") {
  // 重新执行登录流程
  await login();
  // 重试原请求
  return retryRequest();
}
```

---

## 七、数据结构

### 7.1 文件对象

```typescript
interface File {
  id: number;              // 文件ID
  name: string;            // 文件名
  size: number;            // 文件大小（字节）
  lastOpTime: string;      // 最后修改时间 "YYYY-MM-DD HH:mm:ss"
  icon: {
    smallUrl: string;      // 小缩略图URL
    largeUrl?: string;     // 大缩略图URL
  };
  url?: string;            // 文件URL
}
```

### 7.2 文件夹对象

```typescript
interface Folder {
  id: number;              // 文件夹ID
  name: string;            // 文件夹名
  lastOpTime: string;      // 最后修改时间
}
```

### 7.3 文件列表响应

```typescript
interface FilesResponse {
  res_code: number;
  res_message: string;
  fileListAO: {
    count: number;         // 当前页数量
    folderList: Folder[];  // 文件夹列表
    fileList: File[];      // 文件列表
  };
}
```

### 7.4 容量信息

```typescript
interface CapacityInfo {
  res_code: number;
  res_message: string;
  account: string;
  cloudCapacityInfo: {
    freeSize: number;      // 剩余空间
    mail189UsedSize: number;
    totalSize: number;     // 总空间
    usedSize: number;      // 已使用
  };
}
```

---

## 八、使用限制

### 8.1 分页限制

- 最大pageSize: 60
- pageNum从1开始

### 8.2 上传限制

- 默认分片大小: 10MB (10485760字节)
- 分片数量限制:
  - 10MB/20MB分片: 最大999片
  - 50MB以上分片: 最大1999片

### 8.3 并发限制

- 批量任务需要间隔等待
- 删除任务建议间隔200ms
- 移动/复制任务建议间隔400ms

---

## 九、最佳实践

### 9.1 登录持久化

```javascript
// 保存Cookie
const cookies = response.headers['set-cookie'];
saveCookies(cookies);

// 检查登录状态
const testUrl = 'https://cloud.189.cn/api/portal/loginUrl.action';
const response = await request(testUrl);
if (response.redirected && response.url == 'https://cloud.189.cn/web/main') {
  // 已登录
  return true;
}
// 需要重新登录
return false;
```

### 9.2 文件列表遍历

```javascript
async function listAllFiles(folderId) {
  const files = [];
  let pageNum = 1;
  
  while (true) {
    const response = await listFiles(folderId, pageNum);
    if (response.fileListAO.count == 0) {
      break;
    }
    
    files.push(...response.fileListAO.folderList);
    files.push(...response.fileListAO.fileList);
    pageNum++;
  }
  
  return files;
}
```

### 9.3 批量任务等待

```javascript
async function waitBatchTask(type, taskId) {
  while (true) {
    const state = await checkBatchTask(type, taskId);
    switch (state.taskStatus) {
      case 2: // 冲突
        throw new Error('存在冲突');
      case 4: // 完成
        return true;
    }
    await sleep(400); // 等待400ms
  }
}
```

---

## 十、注意事项

1. **登录动态参数**：每次登录的appId、lt、reqId都会变化，需要动态获取
2. **验证码处理**：登录可能触发验证码，需要OCR识别或手动输入
3. **Cookie管理**：使用Cookie维持会话，需要妥善保存和恢复
4. **重定向处理**：下载链接需要多次302重定向，注意提取真实URL
5. **时间格式**：时间均为北京时间格式 "YYYY-MM-DD HH:mm:ss"
6. **不支持家庭云**：Web版API不支持家庭云相关操作

---

## 十一、参考实现

完整的Go语言实现参考：
- 仓库地址：https://github.com/OpenListTeam/OpenList
- 驱动目录：drivers/189/
- 主要文件：
  - driver.go：核心驱动实现
  - login.go：登录流程
  - util.go：上传实现
  - types.go：数据结构定义