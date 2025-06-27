#!/bin/bash

# V2Ray Subscription Manager 构建脚本

set -e

echo "=== V2Ray Subscription Manager 构建脚本 ==="

# 创建输出目录
mkdir -p bin

# 获取版本信息
VERSION=${1:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

echo "版本: ${VERSION}"
echo "构建时间: ${BUILD_TIME}"
echo "Git提交: ${GIT_COMMIT}"

# 构建不同平台的二进制文件
echo "正在构建..."

# Linux amd64
echo "构建 Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-linux-amd64 ./cmd/v2ray-manager

# Linux arm64
echo "构建 Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-linux-arm64 ./cmd/v2ray-manager

# macOS amd64
echo "构建 macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-darwin-amd64 ./cmd/v2ray-manager

# macOS arm64 (Apple Silicon)
echo "构建 macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-darwin-arm64 ./cmd/v2ray-manager

# Windows amd64
echo "构建 Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-windows-amd64.exe ./cmd/v2ray-manager

# Windows arm64
echo "构建 Windows arm64..."
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager-windows-arm64.exe ./cmd/v2ray-manager

# 本地平台构建
echo "构建本地版本..."
go build -ldflags="${LDFLAGS}" -o bin/v2ray-subscription-manager ./cmd/v2ray-manager

echo "✅ 构建完成！"
echo "输出目录: bin/"
ls -la bin/ 