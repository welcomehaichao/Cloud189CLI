# 输出结构详解

## 目录

1. [文件列表输出](#文件列表输出)
2. [容量信息输出](#容量信息输出)
3. [分享输出](#分享输出)
4. [下载链接输出](#下载链接输出)
5. [上传输出](#上传输出)
6. [家庭云列表输出](#家庭云列表输出)
7. [错误输出](#错误输出)

---

## 文件列表输出

```json
{
  "files": [
    {
      "fileId": "123456",
      "fileName": "文档",
      "fileSize": 0,
      "isFolder": true,
      "lastOpTime": "2024-04-01 10:00:00",
      "fileType": null
    },
    {
      "fileId": "789012",
      "fileName": "report.pdf",
      "fileSize": 1536000,
      "isFolder": false,
      "lastOpTime": "2024-04-02 15:30:00",
      "fileType": "pdf"
    }
  ],
  "count": 2
}
```

字段说明：
- `fileId` - 文件唯一标识
- `fileName` - 文件名
- `fileSize` - 文件大小（字节），文件夹为0
- `isFolder` - 是否为文件夹
- `lastOpTime` - 最后修改时间
- `fileType` - 文件类型扩展名

大小转换：
- KB: `fileSize / 1024`
- MB: `fileSize / 1024 / 1024`
- GB: `fileSize / 1024 / 1024 / 1024`

---

## 容量信息输出

```json
{
  "account": "177****8282@189.cn",
  "personalCloud": {
    "totalCapacity": 2247341637632,
    "usedCapacity": 152927506,
    "remainingCapacity": 2247188710126,
    "usageRatio": 0.00007
  },
  "familyCloud": {
    "totalCapacity": 2199123918848,
    "usedCapacity": 62285418,
    "remainingCapacity": 2199061633430,
    "usageRatio": 0.00003
  }
}
```

---

## 分享输出

```json
{
  "shareId": "abc123",
  "shareUrl": "https://cloud.189.cn/web/share?code=abc123",
  "accessCode": "xyz789",
  "expireTime": "2024-04-08 23:59:59",
  "fileName": "report.pdf",
  "fileSize": 1536000
}
```

---

## 下载链接输出

```json
{
  "downloadUrl": "https://download.cloud189.cn/...",
  "expireTime": "2024-04-04 23:59:59",
  "fileName": "report.pdf",
  "fileSize": 1536000
}
```

**重要**：
- 链接有效期约24小时
- 直链，无需登录即可下载
- 过期后需重新调用 `get-url`

---

## 上传输出

```json
{
  "success": true,
  "fileId": "123456",
  "fileName": "uploaded.zip",
  "fileSize": 10485760,
  "filePath": "/autowork/2024-04-04/uploaded.zip"
}
```

---

## 家庭云列表输出

```json
{
  "familyClouds": [
    {
      "familyId": "family001",
      "familyName": "我的家庭",
      "totalCapacity": 2199123918848,
      "usedCapacity": 62285418
    }
  ]
}
```

---

## 错误输出

```json
{
  "error": "error_code",
  "message": "错误描述",
  "details": {}
}
```

常见错误码：
- `not_logged_in` - 未登录
- `session_expired` - Session过期
- `path_not_found` - 路径不存在
- `file_exists` - 文件已存在
- `rate_limit` - API限流
- `upload_failed` - 上传失败