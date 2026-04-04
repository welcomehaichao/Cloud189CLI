# cloud189 CLI Windows 安装脚本
# PowerShell 脚本

param(
    [string]$Version = "latest"
)

$Repo = "welcomehaichao/Cloud189CLI"
$ErrorActionPreference = "Stop"

function Write-Info {
    Write-Host "[INFO] " -ForegroundColor Green -NoNewline
    Write-Host $args
}

function Write-Warn {
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $args
}

function Write-Error {
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $args
}

# 检测架构
function Detect-Platform {
    $Arch = "amd64"
    $BinaryName = "cloud189-windows-$Arch"
    Write-Info "Detected platform: windows/$Arch"
    return $BinaryName
}

# 获取最新版本号
function Get-LatestVersion {
    if ($Version -eq "latest") {
        Write-Info "Fetching latest version..."
        try {
            $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
            $Version = $Release.tag_name
            Write-Info "Latest version: $Version"
        } catch {
            Write-Error "Failed to get latest version"
            exit 1
        }
    }
    return $Version
}

# 下载二进制文件
function Download-Binary {
    $BinaryName = Detect-Platform
    $Version = Get-LatestVersion
    $Url = "https://github.com/$Repo/releases/download/$Version/$BinaryName.zip"
    
    Write-Info "Downloading from: $Url"
    
    # 创建临时目录
    $TmpDir = Join-Path $env:TEMP "cloud189-install"
    New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null
    
    $ZipFile = Join-Path $TmpDir "$BinaryName.zip"
    
    # 下载文件
    try {
        Invoke-WebRequest -Uri $Url -OutFile $ZipFile
        Write-Info "Download completed"
    } catch {
        Write-Error "Download failed: $_"
        exit 1
    }
    
    return @{
        TmpDir = $TmpDir
        ZipFile = $ZipFile
        BinaryName = $BinaryName
    }
}

# 解压并安装
function Install-Binary {
    $DownloadInfo = Download-Binary
    
    Write-Info "Extracting binary..."
    
    # 确定安装目录
    $InstallDir = Join-Path $env:LOCALAPPDATA "cloud189"
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    
    # 解压文件
    try {
        Expand-Archive -Path $DownloadInfo.ZipFile -DestinationPath $InstallDir -Force
    } catch {
        Write-Error "Extraction failed: $_"
        exit 1
    }
    
    # 重命名二进制文件
    $OldBinary = Join-Path $InstallDir "$($DownloadInfo.BinaryName).exe"
    $NewBinary = Join-Path $InstallDir "cloud189.exe"
    
    if (Test-Path $OldBinary) {
        Move-Item -Path $OldBinary -Destination $NewBinary -Force
    }
    
    # 清理临时文件
    Remove-Item -Path $DownloadInfo.TmpDir -Recurse -Force
    
    Write-Info "Installation completed"
    return $InstallDir
}

# 添加到 PATH
function Add-ToPath {
    $InstallDir = Install-Binary
    
    # 检查是否已在 PATH 中
    $CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Info "Adding to PATH..."
        $NewPath = "$CurrentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        
        # 更新当前 session 的 PATH
        $env:PATH += ";$InstallDir"
        
        Write-Info "Added $InstallDir to PATH"
        Write-Warn "You may need to restart PowerShell to use cloud189"
    } else {
        Write-Info "$InstallDir already in PATH"
    }
}

# 验证安装
function Verify-Installation {
    $InstallDir = Join-Path $env:LOCALAPPDATA "cloud189"
    $Binary = Join-Path $InstallDir "cloud189.exe"
    
    if (Test-Path $Binary) {
        Write-Info "Verifying installation..."
        & $Binary version
        Write-Host "✅ cloud189 CLI installed successfully!" -ForegroundColor Green
        Write-Info "Location: $InstallDir"
        
        if (Get-Command cloud189 -ErrorAction SilentlyContinue) {
            Write-Info "cloud189 is available in PATH"
        } else {
            Write-Warn "cloud189 not yet in PATH. Please restart PowerShell or run:"
            Write-Host "    `$env:PATH += `";$InstallDir`"" -ForegroundColor Yellow
        }
    } else {
        Write-Error "Installation failed - binary not found"
        exit 1
    }
}

# 下载并安装 skill
function Download-And-Install-Skill {
    param($targetDir)
    
    $skillUrl = "https://github.com/$Repo/releases/latest/download/cloud189.skill.zip"
    $tmpDir = Join-Path $env:TEMP "cloud189-skill-download"
    $skillZip = Join-Path $tmpDir "cloud189.skill.zip"
    
    New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
    
    Write-Info "下载技能包..."
    
    try {
        Invoke-WebRequest -Uri $skillUrl -OutFile $skillZip
        Write-Info "下载完成"
    } catch {
        Write-Warn "下载失败: $_"
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        return $false
    }
    
    Write-Info "安装技能包..."
    
    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Force -Path $targetDir | Out-Null
    }
    
    try {
        Expand-Archive -Path $skillZip -DestinationPath $targetDir -Force
        Write-Info "✓ Skill 已安装到: $targetDir\cloud189"
        Write-Info "请重启 Agent 工具"
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        return $true
    } catch {
        Write-Warn "解压失败: $_"
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        return $false
    }
}

# 检测并自动安装 AI Agent skill
function Install-Skill {
    Write-Info "检测 AI Agent 工具..."
    
    $agentDirs = @(
        Join-Path $env:USERPROFILE ".claude\skills"
        Join-Path $env:APPDATA "opencode\skills"
        Join-Path $env:USERPROFILE ".openclaw\skills"
    )
    
    $installedCount = 0
    
    foreach ($agentDir in $agentDirs) {
        if (Test-Path $agentDir) {
            $agentName = Split-Path (Split-Path $agentDir) -Leaf
            
            Write-Info "发现 $agentName 工具"
            
            $skillPath = Join-Path $agentDir "cloud189"
            
            if (Test-Path $skillPath) {
                Write-Info "cloud189 skill 已存在，跳过"
                continue
            }
            
            # 自动安装（无需询问）
            if (Download-And-Install-Skill $agentDir) {
                $installedCount++
            }
        fi
    }
    
    if ($installedCount -gt 0) {
        Write-Info "✓ 已为 $installedCount 个 Agent 工具安装 skill"
    } else {
        Write-Info "未检测到 AI Agent 工具"
        Write-Info "如需手动安装，请从 GitHub Release 下载 cloud189.skill.zip"
    }
}

# 主流程
Write-Info "Starting cloud189 CLI installation..."

Add-ToPath
Verify-Installation

# 自动安装 AI Agent skill（无需询问）
Install-Skill

Write-Info "Installation completed!"