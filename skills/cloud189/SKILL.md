---
name: cloud189
description: |
  天翼云盘CLI工具，用于管理天翼云盘文件。当用户提到天翼云盘、Cloud189、cloud189、189云盘、电信云盘时必须使用此skill。当用户需要：上传/下载云端文件、管理云端文件夹、创建文件分享链接、查看云端容量、操作家庭云、获取下载链接、备份文件到云端、从云端恢复文件、发送文件给他人时，必须使用此skill。特别触发场景：用户说"把文件发给我"、"上传并给我链接"、"发送这个文件"、"给我下载链接"等。即使用户只说"帮我上传文件"或"分享这个文件"未明确提及天翼云盘，只要上下文涉及云盘操作，也应主动询问是否使用天翼云盘并触发此skill。
allowed-tools:
  - Bash(cloud189 *)
---

# 天翼云盘 CLI 工具

## 前置检查

**重要**：工具内置help，对功能不清楚时**优先查询help**：

```bash
cloud189 --help           # 查看所有命令
cloud189 <命令> --help    # 查看命令详情，如 cloud189 upload --help
```

help信息最准确、最新，优先于参考文档。

```bash
cloud189 version  # 确认已安装
cloud189 whoami   # 确认登录状态
cloud189 login --qr  # 未登录时使用二维码登录
```

## 快速参考

| 操作 | 命令 |
|------|------|
| 列出 | `cloud189 ls [路径]` |
| 上传 | `cloud189 upload <本地> <云端>` |
| 下载 | `cloud189 download <云端> <本地>` |
| 下载链接 | `cloud189 get-url <路径>` |
| 分享 | `cloud189 share <路径> --expire 7d --code` |
| 容量 | `cloud189 info` |
| 家庭云 | `cloud189 family list` |

详细命令参考：[references/commands.md](references/commands.md)

## 输出格式

默认JSON输出，使用 `-o json|yaml|table` 切换。

输出结构详解：[references/output-structures.md](references/output-structures.md)

## 核心工作流

### 发送文件给用户（上传+生成下载链接）

用户说"把文件发给我"、"给我下载链接"时使用。

目标路径：`/autowork/{YYYY-MM-DD}/` 按日期归档。

```bash
# 1. 确认登录
cloud189 whoami

# 2. 创建日期目录
cloud189 mkdir /autowork/2024-04-04

# 3. 上传文件
cloud189 upload ./file.zip /autowork/2024-04-04/

# 4. 生成下载链接
cloud189 get-url /autowork/2024-04-04/file.zip -o json
```

返回给用户：
- `downloadUrl` - 直接下载链接（无需登录）
- `expireTime` - 过期时间（约24小时）

### 上传文件

```bash
cloud189 mkdir /备份/2024-04
cloud189 upload ./file.zip /备份/2024-04/
cloud189 upload ./large.zip / --stream --resume  # 大文件+断点续传
```

### 分享文件（长期分享）

```bash
cloud189 share /文档/report.pdf --expire 7d --code -o json
# 返回 shareUrl + accessCode
```

### 家庭云操作

```bash
cloud189 family list
cloud189 family use <ID>
cloud189 ls / --family
cloud189 family save /视频/xxx.mp4 /我的视频/
```

详细工作流参考：[references/commands.md](references/commands.md)

## 错误处理

| 错误 | 处理 |
|------|------|
| 未登录 | `cloud189 login --qr` |
| 路径不存在 | 先 `ls` 确认，再 `mkdir` |
| 上传失败 | `--stream --resume` 断点续传 |
| API限流 | 降低频率，添加延迟 |

详细错误处理：[references/error-handling.md](references/error-handling.md)

## 安装

Linux/macOS:
```bash
curl -sL https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.sh | bash
```

Windows:
```powershell
Invoke-WebRequest -Uri "https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.ps1" | Invoke-Expression
```

## Reference Files

- `references/commands.md` - 详细命令参考（认证、文件操作、上传下载、分享、家庭云）
- `references/output-structures.md` - JSON输出结构详解
- `references/error-handling.md` - 错误处理指南