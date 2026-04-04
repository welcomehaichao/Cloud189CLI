# 错误处理指南

## 常见错误及处理策略

| 错误类型 | 错误信息 | 处理方式 |
|----------|----------|----------|
| 认证错误 | `not_logged_in` | 执行 `cloud189 login --qr` |
| 认证错误 | `session_expired` | 自动刷新，失败则重新登录 |
| 路径错误 | `path_not_found` | 先 `ls` 确认，必要时 `mkdir` |
| 文件错误 | `file_exists` | 提示用户，询问是否覆盖/重命名 |
| 上传错误 | `upload_failed` | 使用 `--stream --resume` 断点续传 |
| 下载错误 | `download_failed` | 使用 `--resume` 断点续传 |
| 限流错误 | `rate_limit` | 降低频率，添加延迟 |
| 网络错误 | `network_error` | 检查网络，重试操作 |

## 错误处理流程

1. **识别错误类型**
   - 解析JSON输出的 `error` 字段
   - 匹配对应的处理策略

2. **执行修复操作**
   - 认证问题：重新登录
   - 路径问题：创建或确认路径
   - 传输问题：启用断点续传
   - 限流问题：添加延迟

3. **重试原操作**
   - 最多重试3次
   - 每次重试间隔递增（1s → 2s → 4s）

## 断点续传使用

上传断点续传：
```bash
cloud189 upload ./large.zip / --stream --resume
```

下载断点续传：
```bash
cloud189 download /视频/movie.mp4 ./ --resume
```

进度自动保存，中断后可继续。

## 二维码登录问题

二维码无法识别时：
1. 确保版本 ≥ v1.0.1
2. 使用等宽字体（Consolas/Monaco）
3. 终端宽度 ≥ 80列
4. 使用备用链接在浏览器打开

备用方案：密码登录
```bash
cloud189 login -u <用户名> -p <密码>
```

## 速率限制处理

批量操作建议：
- 文件操作串行执行
- 每次操作间隔1-2秒
- 使用 `sleep` 控制频率

```bash
cloud189 upload ./1.zip /
sleep 2
cloud189 upload ./2.zip /
```