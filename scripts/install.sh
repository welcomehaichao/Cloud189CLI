#!/bin/bash

# cloud189 CLI 安装脚本
# 支持 Linux 和 macOS

set -e

VERSION="${1:-latest}"
REPO="welcomehaichao/Cloud189CLI"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检测操作系统和架构
detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case "$OS" in
        Darwin)
            OS_NAME="darwin"
            ;;
        Linux)
            OS_NAME="linux"
            ;;
        *)
            echo_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH_NAME="amd64"
            ;;
        aarch64|arm64)
            ARCH_NAME="arm64"
            ;;
        *)
            echo_error "Unsupported ARCH: $ARCH"
            exit 1
            ;;
    esac
    
    BINARY_NAME="cloud189-${OS_NAME}-${ARCH_NAME}"
    echo_info "Detected platform: ${OS_NAME}/${ARCH_NAME}"
}

# 获取最新版本号
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        echo_info "Fetching latest version..."
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo_error "Failed to get latest version"
            exit 1
        fi
        echo_info "Latest version: ${VERSION}"
    fi
}

# 下载二进制文件
download_binary() {
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"
    
    echo_info "Downloading from: ${URL}"
    
    # 创建临时目录
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    # 下载文件
    if command -v wget >/dev/null 2>&1; then
        wget -q "$URL" -O "${BINARY_NAME}.tar.gz"
    elif command -v curl >/dev/null 2>&1; then
        curl -sL "$URL" -o "${BINARY_NAME}.tar.gz"
    else
        echo_error "wget or curl required"
        exit 1
    fi
    
    # 检查文件是否下载成功
    if [ ! -f "${BINARY_NAME}.tar.gz" ]; then
        echo_error "Download failed"
        exit 1
    fi
    
    echo_info "Download completed"
}

# 解压并安装
install_binary() {
    echo_info "Extracting binary..."
    tar -xzf "${BINARY_NAME}.tar.gz"
    
    # 检查解压后的文件
    if [ ! -f "$BINARY_NAME" ]; then
        echo_error "Binary not found after extraction"
        exit 1
    fi
    
    # 添加执行权限
    chmod +x "$BINARY_NAME"
    
    # 确定安装路径
    INSTALL_DIR="/usr/local/bin"
    
    # 检查是否有写入权限
    if [ ! -w "$INSTALL_DIR" ]; then
        echo_warn "No write permission to ${INSTALL_DIR}, using ~/bin instead"
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
        
        # 添加到 PATH（如果需要）
        if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
            echo_warn "Adding $INSTALL_DIR to PATH..."
            
            # 根据使用的 shell 添加到配置文件
            if [ -f "$HOME/.bashrc" ]; then
                echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
                echo_info "Added to ~/.bashrc. Run 'source ~/.bashrc' to update PATH"
            fi
            
            if [ -f "$HOME/.zshrc" ]; then
                echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc"
                echo_info "Added to ~/.zshrc. Run 'source ~/.zshrc' to update PATH"
            fi
        fi
    fi
    
    # 安装二进制文件
    echo_info "Installing to ${INSTALL_DIR}/cloud189"
    
    if [ -w "/usr/local/bin" ]; then
        sudo mv "$BINARY_NAME" "${INSTALL_DIR}/cloud189"
    else
        mv "$BINARY_NAME" "${INSTALL_DIR}/cloud189"
    fi
    
    # 清理临时目录
    cd - > /dev/null
    rm -rf "$TMP_DIR"
    
    echo_info "Installation completed"
}

# 验证安装
verify_installation() {
    if command -v cloud189 >/dev/null 2>&1; then
        echo_info "Verifying installation..."
        cloud189 version
        echo -e "${GREEN}✅ cloud189 CLI installed successfully!${NC}"
    else
        echo_warn "cloud189 not found in PATH. You may need to restart your terminal or run:"
        echo "    source ~/.bashrc  # for bash"
        echo "    source ~/.zshrc   # for zsh"
        echo "Or manually add to PATH:"
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# 主流程
main() {
    echo_info "Starting cloud189 CLI installation..."
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
}

main