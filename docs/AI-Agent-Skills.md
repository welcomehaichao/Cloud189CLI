# AI Agent 技能包使用指南

本项目为 AI Agent 工具提供预配置的技能包，让 AI 助手能够直接操作天翼云盘。

## 下载技能包

从 [GitHub Releases](https://github.com/welcomehaichao/Cloud189CLI/releases) 下载 `cloud189.skill.zip`。

## 在 Agent 工具中使用

### Claude Code

Claude Code 是 Anthropic 官方的 AI 编程助手工具。

```bash
# 1. 下载技能包
wget https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189.skill.zip

# 2. 解压到 Claude Code 技能目录
unzip cloud189.skill.zip -d ~/.claude/skills/

# 3. 重启 Claude Code 或重新加载配置
# 技能将自动加载，AI 可以直接使用 cloud189 命令
```

### OpenCode

OpenCode 是开源的 AI 编程助手。

```bash
# 1. 下载技能包
wget https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189.skill.zip

# 2. 解压到 OpenCode 技能目录
unzip cloud189.skill.zip -d ~/.config/opencode/skills/

# 3. 重启 OpenCode
# 在对话中提到"天翼云盘"或相关操作时，AI 会自动使用此技能
```

### OpenClaw

OpenClaw 是另一个流行的 AI Agent 工具。

```bash
# 1. 下载技能包
wget https://github.com/welcomehaichao/Cloud189CLI/releases/latest/download/cloud189.skill.zip

# 2. 解压到 OpenClaw 技能目录
unzip cloud189.skill.zip -d ~/.openclaw/skills/

# 3. 重启 OpenClaw
# AI 助手将能够识别并执行天翼云盘相关操作
```

## 技能包内容

技能包包含以下内容：

```
cloud189/
├── SKILL.md                    # 技能主文件（触发规则和使用指南）
└── references/                 # 参考文档
    ├── commands.md             # 详细命令参考
    ├── output-structures.md    # JSON 输出结构说明
    └── error-handling.md       # 错误处理指南
```

## 技能触发场景

当用户提到以下关键词时，AI Agent 会自动触发此技能：

- **云盘名称**: 天翼云盘、Cloud189、cloud189、189云盘、电信云盘
- **操作需求**: 上传/下载文件、创建分享链接、查看容量、操作家庭云
- **特定场景**: "把文件发给我"、"上传并给我链接"、"发送这个文件"、"给我下载链接"

## 技能功能

- ✅ 自动识别天翼云盘相关请求
- ✅ 智能选择合适的命令执行
- ✅ 解析 JSON 输出并提取关键信息
- ✅ 错误处理和重试机制
- ✅ 支持大文件断点续传

## 技能包自动打包

每次发布新版本时，GitHub Actions 会自动打包技能包：

1. 遍历 `skills/` 目录下的所有技能
2. 将每个技能打包为 `{技能名}.skill.zip`
3. 上传到 GitHub Release 作为发布资产

## 技能开发

如需开发新的技能包，请参考以下结构：

```
skills/
└── your-skill/
    ├── SKILL.md                # 必需：技能主文件
    └── references/             # 可选：参考文档目录
        ├── commands.md
        ├── output-structures.md
        └── error-handling.md
```

SKILL.md 必须包含 YAML front matter：

```yaml
---
name: your-skill
description: 技能描述，用于触发识别
allowed-tools:
  - Bash(command *)
---
```

## 常见问题

### 技能未加载？

1. 确认技能包已解压到正确的技能目录
2. 检查 SKILL.md 文件是否存在
3. 重启 Agent 工具或重新加载配置

### AI 未识别天翼云盘操作？

1. 在对话中明确提及"天翼云盘"、"Cloud189"等关键词
2. 使用具体操作描述，如"上传文件到天翼云盘"
3. 检查技能包的 description 是否包含相关关键词

### 更多信息

- [项目 README](../README.md)
- [技能包源码](../skills/cloud189/)
- [GitHub Releases](https://github.com/welcomehaichao/Cloud189CLI/releases)