# 天翼云盘 CLI 工具

一个功能完整的天翼云盘命令行工具，专为 AI Agent 调用和自动化脚本设计。

## 功能特性

✅ **认证方式**: 支持密码登录和二维码登录  
✅ **Token持久化**: 自动保存和刷新登录状态  
✅ **文件操作**: 列表、创建、删除、移动、复制、重命名  
✅ **家庭云支持**: 完整支持个人云和家庭云操作  
✅ **机器可读输出**: 支持 JSON/YAML 格式输出，便于 AI 解析  
✅ **AI Agent技能包**: 提供开箱即用的技能包，支持主流 Agent 工具

## AI Agent 技能包

本项目提供开箱即用的技能包，支持主流 AI Agent 工具（Claude Code、OpenCode、OpenClaw）。

从 [GitHub Releases](https://github.com/welcomehaichao/Cloud189CLI/releases) 下载 `cloud189.skill.zip`，解压到对应工具的技能目录即可使用。

详细使用指南请查看：**[AI Agent 技能包使用指南](docs/AI-Agent-Skills.md)**

## 安装

### 一键安装（推荐）

#### Linux / macOS

```bash
# 使用 curl
curl -sL https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.sh | bash

# 或使用 wget
wget -qO- https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.sh | bash
```

#### Windows (PowerShell)

```powershell
# 一键安装
Invoke-WebRequest -Uri "https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.ps1" | Invoke-Expression
```

### 手动下载

从 [GitHub Releases](https://github.com/welcomehaichao/Cloud189CLI/releases) 下载对应平台的二进制文件：

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | x64 | `cloud189-linux-amd64.tar.gz` |
| Linux | ARM64 | `cloud189-linux-arm64.tar.gz` |
| macOS | Intel | `cloud189-darwin-amd64.tar.gz` |
| macOS | M1/M2 | `cloud189-darwin-arm64.tar.gz` |
| Windows | x64 | `cloud189-windows-amd64.zip` |

#### Linux / macOS 手动安装

```bash
# 下载
wget https://github.com/welcomehaichao/Cloud189CLI/releases/download/v1.2.0/cloud189-linux-amd64.tar.gz

# 解压
tar -xzf cloud189-linux-amd64.tar.gz

# 安装
chmod +x cloud189-linux-amd64
sudo mv cloud189-linux-amd64 /usr/local/bin/cloud189

# 验证
cloud189 version
```

#### Windows 手动安装

```powershell
# 下载
Invoke-WebRequest -Uri "https://github.com/welcomehaichao/Cloud189CLI/releases/download/v1.2.0/cloud189-windows-amd64.zip" -OutFile "cloud189.zip"

# 解压
Expand-Archive -Path "cloud189.zip" -DestinationPath "C:\cloud189"

# 添加到 PATH（可选）
$env:PATH += ";C:\cloud189"

# 验证
C:\cloud189\cloud189.exe version
```

## 更新

### 一键更新

已安装用户可直接重新运行安装脚本覆盖更新：

#### Linux / macOS

```bash
curl -sL https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.sh | bash
```

#### Windows (PowerShell)

```powershell
iwr -useb https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189-install.ps1 | iex
```

### 查看当前版本

```bash
cloud189 version
```

### 更新提示

- 安装脚本默认下载最新版本
- 更新会自动覆盖旧版本，无需手动卸载
- 更新后可能需要重启终端使PATH生效

## 快速开始

### 编译

```bash
go build -o cloud189 ./cmd/cloud189
```

### 登录

```bash
# 密码登录
cloud189 login -u 用户名 -p 密码

# 二维码登录（推荐）
cloud189 login --qr

# 查看登录状态
cloud189 whoami

# 退出登录
cloud189 logout
```

#### 二维码登录特性
- ✅ **终端直接显示二维码** - 无需浏览器，直接在命令行看到二维码
- ✅ **实时状态提示** - 显示扫码和确认状态
- ✅ **友好交互体验** - 清晰的视觉反馈
- ✅ **高容错级别** - 使用High级别容错（30%恢复能力），提高识别率

#### 二维码识别说明

**重要**: 确保使用 **v1.0.1+** 版本，已修复定位角缺失问题。

二维码有三个定位角（左上、右上、左下），每个都是清晰的 **大方块**。如果看不到定位角，请：
1. 确保终端窗口足够大
2. 使用等宽字体（如 Consolas）
3. 不要缩放终端窗口

示例输出：
```
========== 天翼云盘二维码登录 ==========

████████████████████████████████████
████████████████████████████████████
██                              ██
██  ████████  ██    ████████  ██
██  ██    ████  ██  ██    ████
...
████████████████████████████████████

或直接访问: https://m.cloud.189.cn/zhuanti/qrLogin/...

请使用天翼云盘APP扫描上方二维码
======================================

⏳ 等待扫码...
```

### 文件操作

```bash
# 列出文件
cloud189 ls /                    # 根目录
cloud189 ls /文档                # 绝对路径
cloud189 ls test01               # 相对路径
cloud189 ls /文档 --long         # 详细列表
cloud189 ls / --output yaml     # YAML格式
cloud189 ls / --output table    # 表格格式
cloud189 ls / --family          # 家庭云

# 创建文件夹
cloud189 mkdir 新文件夹          # 在根目录创建
cloud189 mkdir /文档/新文件夹    # 在指定路径创建
cloud189 mkdir /文档 --family   # 在家庭云创建

# 查看容量
cloud189 info
cloud189 info -o table          # 表格格式
```

### 家庭云

```bash
# 列出家庭云
cloud189 family list

# 切换家庭云
cloud189 family use <家庭云ID>

# 家庭云文件列表
cloud189 ls / --family

# 家庭云文件转存到个人云
cloud189 family save <家庭云文件路径> <个人云路径>
```

### 文件分享

```bash
# 创建分享链接
cloud189 share <文件路径>
cloud189 share /文档/test.txt --expire 7d  # 7天有效期
cloud189 share /文档/test.txt --code       # 生成提取码

# 列出我的分享
cloud189 share-list

# 取消分享
cloud189 share-cancel <分享ID>
```

### 获取下载链接

```bash
# 获取文件下载链接
cloud189 get-url <文件路径>
cloud189 get-url /文档/test.txt
cloud189 get-url /视频/movie.mp4 --family
```

### 日志管理

```bash
# 查看最近日志
cloud189 log view

# 查看日志统计
cloud189 log stats

# 查看日志信息
cloud189 log info

# 清理过期日志
cloud189 log clean
```

## 命令详解

### 全局选项

```
-o, --output string   输出格式 (json|yaml|table) (默认: json)
-h, --help           显示帮助信息
```

### 登录命令

```bash
cloud189 login [选项]

选项:
  -u, --username string   用户名
  -p, --password string   密码
  -q, --qr               二维码登录
```

### 文件列表

```bash
cloud189 ls [路径] [选项]

选项:
  -l, --long             显示详细信息
  -r, --recursive        递归列出
      --order-by string  排序字段 (filename|filesize|lastOpTime)
      --desc            降序排列
  -n, --page int        页码 (默认: 1)
      --page-size int   每页数量 (默认: 100)
      --family          家庭云
```

### 创建文件夹

```bash
cloud189 mkdir <路径> [选项]

选项:
  --family   家庭云
```

### 容量信息

```bash
cloud189 info
```

### 家庭云操作

```bash
cloud189 family list    # 列出家庭云
```

## 输出格式

支持三种输出格式：

```bash
-o json    # JSON格式（默认）
-o yaml    # YAML格式
-o table   # 表格格式（人类可读）
```

### 表格格式示例

查看容量信息：
```bash
.\cloud189.exe info -o table
```

输出：
```
账号: 177****8282@189.cn

=== 个人云 ===
类型               大小              GB
-------------------------------------------------------
总容量           2247341637632         2093.00
已使用             152927506            0.14
剩余空间          2247188710126         2092.86
使用率                               0.0%

=== 家庭云 ===
类型               大小              GB
-------------------------------------------------------
总容量           2199123918848         2048.09
已使用              62285418            0.06
剩余空间          2199061633430         2048.03
使用率                               0.0%
```

查看文件列表：
```bash
.\cloud189.exe ls / -o table
```

输出：
```
名称                 大小        类型            修改时间
----------------------------------------------------------------------
文档                 -          文件夹          2024-03-31 10:00:00
照片                 -          文件夹          2024-03-30 15:30:00
test.txt            1.5 KB      文件           2024-03-31 12:00:00
```

### YAML 格式

```bash
cloud189 ls / --output yaml
```

## 项目结构

```
Cloud189CLI/
├── cmd/cloud189/          # 主程序入口
├── internal/
│   ├── api/              # API 客户端
│   │   ├── client.go     # HTTP 客户端
│   │   ├── auth.go       # 认证 API
│   │   ├── file.go       # 文件操作 API
│   │   ├── batch.go      # 批量任务 API
│   │   └── family.go     # 家庭云 API
│   ├── crypto/           # 加密算法
│   ├── commands/         # CLI 命令
│   ├── config/           # 配置管理
│   └── output/           # 输出格式
├── pkg/
│   ├── types/            # 数据类型
│   └── utils/            # 工具函数
└── docs/                 # API 文档
```

## 开发计划

详细的功能清单请查看：**[功能清单.md](功能清单.md)**

### 当前进度
- ✅ 已完成：40项功能
- ⬜ 待实现：4项功能
- 📊 完成率：91%

### 已完成核心功能
- ✅ 文件删除 (`rm`)
- ✅ 文件移动 (`mv`)
- ✅ 文件复制 (`cp`)
- ✅ 文件重命名 (`rename`)
- ✅ 文件下载 (`download`)
- ✅ 文件分享 (`share`, `share-list`, `share-cancel`)
- ✅ 家庭云切换 (`family use`)
- ✅ 家庭云转存 (`family save`)
- ✅ Session自动刷新
- ✅ 配置加密存储
- ✅ 日志审计系统

### 优先实现（P1级别）
- [ ] 回收站管理
- [ ] 验证码自动识别

完整功能清单和开发进度请查看 [功能清单.md](功能清单.md)。

## 技术栈

- **语言**: Go 1.23+
- **CLI框架**: [Cobra](https://github.com/spf13/cobra)
- **HTTP客户端**: [Resty](https://github.com/go-resty/resty)
- **加密**: Go标准库 crypto/*

## 配置文件

配置文件位于 `~/.cloud189/config.json`，包含：

- 用户名
- Session密钥
- Token信息

## 常见问题

### 二维码无法识别？

1. **确认版本**: 确保使用 v1.0.1+ 版本（已修复定位角问题）
2. **终端设置**: 
   - 使用等宽字体（Consolas / Monaco / Courier New）
   - 字号 14-16pt
   - 终端宽度至少 80 列
3. **扫码技巧**:
   - 保持手机与屏幕 10-20cm 距离
   - 确保光线充足
   - 等待相机自动对焦
4. **备用方案**: 
   - 使用终端显示的链接在浏览器打开
   - 或使用密码登录: `cloud189 login -u 用户名 -p 密码`

详细说明请参考: [二维码登录使用指南](docs/二维码登录使用指南.md)

### 登录失败？

1. 检查网络连接
2. 确认用户名密码正确
3. 尝试二维码登录
4. 删除配置文件重新登录: `rm ~/.cloud189/config.json`

## API 文档

详细的天翼云盘 API 文档请参考：

- [天翼云盘Web版API文档](./docs/天翼云盘Web版API文档.md)
- [天翼云盘PC客户端API文档](./docs/天翼云盘PC客户端API文档.md)
- [天翼云盘TV电视端API文档](./docs/天翼云盘TV电视端API文档.md)

## 注意事项

1. 本工具仅用于个人学习和研究目的
2. 请遵守天翼云盘服务条款
3. 不要频繁调用API，以免被限流
4. Token会自动保存在本地，下次使用无需重新登录

## License

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！