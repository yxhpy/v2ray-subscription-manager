#!/bin/bash

# Web UI 构建脚本

set -e

echo "🔧 构建 V2Ray Web UI..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 环境未安装"
    exit 1
fi

# 获取版本信息
VERSION=${1:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

echo "版本: ${VERSION}"
echo "构建时间: ${BUILD_TIME}"
echo "Git提交: ${GIT_COMMIT}"

# 创建输出目录
mkdir -p bin

# 构建不同平台的版本
echo "📦 构建多平台版本..."

# Linux amd64
echo "  🐧 构建 Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-linux-amd64 ./cmd/web-ui/

# Linux arm64
echo "  🐧 构建 Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-linux-arm64 ./cmd/web-ui/

# Windows amd64
echo "  🪟 构建 Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-windows-amd64.exe ./cmd/web-ui/

# Windows arm64
echo "  🪟 构建 Windows arm64..."
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-windows-arm64.exe ./cmd/web-ui/

# macOS amd64
echo "  🍎 构建 macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-darwin-amd64 ./cmd/web-ui/

# macOS arm64 (Apple Silicon)
echo "  🍎 构建 macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-webui-${VERSION}-darwin-arm64 ./cmd/web-ui/

echo ""
echo "✅ 构建完成！"
echo ""
echo "📁 输出文件："
ls -la bin/v2ray-webui-${VERSION}-*

echo ""
echo "🚀 运行方式："
echo "  Linux/macOS: ./bin/v2ray-webui-<platform> [port]"
echo "  Windows:     ./bin/v2ray-webui-<platform>.exe [port]"
echo ""
echo "  默认端口: 8888"
echo "  访问地址: http://localhost:8888"
echo ""
echo "📝 示例："
echo "  ./bin/v2ray-webui-linux-amd64 8080"
echo "  ./bin/v2ray-webui-windows-amd64.exe 9999" 