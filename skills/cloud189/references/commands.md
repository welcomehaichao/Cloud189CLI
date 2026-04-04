# 命令详细参考

## 目录

1. [认证命令](#认证命令)
2. [文件操作命令](#文件操作命令)
3. [上传下载命令](#上传下载命令)
4. [分享命令](#分享命令)
5. [家庭云命令](#家庭云命令)
6. [信息查询命令](#信息查询命令)
7. [日志命令](#日志命令)

---

## 认证命令

### login - 登录

```bash
cloud189 login -u <用户名> -p <密码>  # 密码登录
cloud189 login --qr                    # 二维码登录（推荐）
```

二维码登录特性：
- 终端直接显示ASCII二维码
- 实时状态提示
- 使用High容错级别

### logout - 退出登录

```bash
cloud189 logout  # 清除本地Token
```

### whoami - 查看登录状态

```bash
cloud189 whoami  # 显示当前登录用户
```

---

## 文件操作命令

### ls - 列出文件

```bash
cloud189 ls [路径]
```

参数：
- `-l, --long` - 显示详细信息
- `-r, --recursive` - 递归列出
- `--order-by string` - 排序字段 (filename|filesize|lastOpTime)
- `--desc` - 降序排列
- `-n, --page int` - 页码 (默认: 1)
- `--page-size int` - 每页数量 (默认: 100)
- `--family` - 家庭云

### mkdir - 创建文件夹

```bash
cloud189 mkdir <路径> [--family]
```

支持多级路径创建。

### rm - 删除文件

```bash
cloud189 rm <路径> [--family]
```

移入回收站，非永久删除。

### mv - 移动文件

```bash
cloud189 mv <源路径> <目标路径> [--family]
```

### cp - 复制文件

```bash
cloud189 cp <源路径> <目标路径> [--family]
```

### rename - 重命名

```bash
cloud189 rename <路径> <新名称> [--family]
```

---

## 上传下载命令

### upload - 上传文件

```bash
cloud189 upload <本地文件> <云端路径>
```

参数：
- `--stream` - 分片上传（适合大文件）
- `--resume` - 断点续传

上传模式：
- 默认模式：适合小文件
- Stream模式：适合大文件，自动分片（10MB/20MB/50MB+）

### download - 下载文件

```bash
cloud189 download <云端路径> <本地路径> [--resume]
```

支持断点续传。

### get-url - 获取下载链接

```bash
cloud189 get-url <云端路径> [--family]
```

返回带过期时间的直接下载URL。

---

## 分享命令

### share - 创建分享

```bash
cloud189 share <文件路径>
```

参数：
- `--expire string` - 有效期 (如 7d)
- `--code` - 生成提取码

### share-list - 列出分享

```bash
cloud189 share-list
```

### share-info - 查看分享详情

```bash
cloud189 share-info <分享ID>
```

### share-cancel - 取消分享

```bash
cloud189 share-cancel <分享ID>
```

---

## 家庭云命令

### family list - 列出家庭云

```bash
cloud189 family list
```

### family use - 切换家庭云

```bash
cloud189 family use <家庭云ID>
```

### family save - 家庭云转存

```bash
cloud189 family save <家庭云文件路径> <个人云路径>
```

将家庭云文件保存到个人云。

---

## 信息查询命令

### info - 查看容量

```bash
cloud189 info
```

显示个人云和家庭云容量信息。

---

## 日志命令

### log view - 查看日志

```bash
cloud189 log view
```

### log stats - 日志统计

```bash
cloud189 log stats
```

### log info - 日志信息

```bash
cloud189 log info
```

### log clean - 清理日志

```bash
cloud189 log clean
```

清理180天前的日志。