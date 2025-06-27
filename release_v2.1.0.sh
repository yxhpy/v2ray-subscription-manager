#!/bin/bash

# V2Ray Subscription Manager v2.1.0 发布脚本
# 简化版本，用于手动发布

set -e

VERSION="v2.1.0"
COMMIT_MSG="代理管理器和MVP测试器优化版本 - 增强超时处理、配置展示和Windows平台并发控制"

echo "🚀 开始发布 $VERSION 版本"

# 1. 创建bin目录
mkdir -p bin

# 2. 编译各平台版本
echo "📦 编译各平台版本..."

# Windows amd64
echo "编译 Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-windows-amd64.exe ./cmd/v2ray-manager

# Windows arm64
echo "编译 Windows arm64..."
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-windows-arm64.exe ./cmd/v2ray-manager

# Linux amd64
echo "编译 Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-linux-amd64 ./cmd/v2ray-manager

# Linux arm64
echo "编译 Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-linux-arm64 ./cmd/v2ray-manager

# macOS amd64
echo "编译 macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-darwin-amd64 ./cmd/v2ray-manager

# macOS arm64
echo "编译 macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/v2ray-subscription-manager-darwin-arm64 ./cmd/v2ray-manager

# 3. 创建压缩包
echo "📦 创建压缩包..."
cd bin

# 创建各平台压缩包
zip -q v2ray-subscription-manager-${VERSION}-windows-amd64.zip v2ray-subscription-manager-windows-amd64.exe
zip -q v2ray-subscription-manager-${VERSION}-windows-arm64.zip v2ray-subscription-manager-windows-arm64.exe
zip -q v2ray-subscription-manager-${VERSION}-linux-amd64.zip v2ray-subscription-manager-linux-amd64
zip -q v2ray-subscription-manager-${VERSION}-linux-arm64.zip v2ray-subscription-manager-linux-arm64
zip -q v2ray-subscription-manager-${VERSION}-darwin-amd64.zip v2ray-subscription-manager-darwin-amd64
zip -q v2ray-subscription-manager-${VERSION}-darwin-arm64.zip v2ray-subscription-manager-darwin-arm64

# 创建全平台压缩包
tar -czf v2ray-subscription-manager-${VERSION}-all-platforms.tar.gz v2ray-subscription-manager-*

# 生成校验和
sha256sum *${VERSION}* > v2ray-subscription-manager-${VERSION}-checksums.txt

cd ..

# 4. Git操作
echo "📝 提交代码..."
git add .
git commit -m "release: $COMMIT_MSG

Version: $VERSION
Build Date: $(date '+%Y-%m-%d %H:%M:%S')
Platforms: Windows, Linux, macOS (amd64, arm64)

Changes in this release:
- 优化代理管理器的超时处理机制
- 新增配置展示功能
- 增加用户配置支持
- 改进MVP测试器
- Windows平台并发控制优化
- 连接稳定性提升

Files included:
$(cd bin && ls -1 *${VERSION}* | sed 's/^/- /')"

echo "🏷️ 创建标签..."
git tag -a "$VERSION" -m "Release $VERSION

Build Date: $(date '+%Y-%m-%d %H:%M:%S')
Commit: $(git rev-parse HEAD)

$COMMIT_MSG"

echo "⬆️ 推送到远程..."
git push origin main
git push origin "$VERSION"

# 5. 生成Release说明
cat > release_notes_${VERSION}.md << EOF
# $VERSION - 代理管理器和MVP测试器优化版本

## 🚀 主要更新

### 核心功能增强
- **超时处理增强**: 优化代理管理器的超时处理机制，提供更稳定的连接管理
- **配置展示功能**: 新增配置展示功能，提供更清晰的代理状态信息
- **用户配置支持**: 增加超时时间和测试URL的用户配置支持，提升使用灵活性

### Windows平台专项优化
- **并发控制改进**: 改进Windows平台的并发控制机制，避免竞态条件
- **连接稳定性**: 实现智能URL选择策略，提升连接稳定性

## 📦 下载文件

| 平台 | 架构 | 文件名 | 说明 |
|------|------|--------|------|
| **Windows** | x64 | \`v2ray-subscription-manager-${VERSION}-windows-amd64.zip\` | Windows 64位版本 |
| **Windows** | ARM64 | \`v2ray-subscription-manager-${VERSION}-windows-arm64.zip\` | Windows ARM64版本 |
| **Linux** | x64 | \`v2ray-subscription-manager-${VERSION}-linux-amd64.zip\` | Linux 64位版本 |
| **Linux** | ARM64 | \`v2ray-subscription-manager-${VERSION}-linux-arm64.zip\` | Linux ARM64版本 |
| **macOS** | Intel | \`v2ray-subscription-manager-${VERSION}-darwin-amd64.zip\` | macOS Intel版本 |
| **macOS** | Apple Silicon | \`v2ray-subscription-manager-${VERSION}-darwin-arm64.zip\` | macOS M1/M2版本 |
| **All Platforms** | 通用 | \`v2ray-subscription-manager-${VERSION}-all-platforms.tar.gz\` | 所有平台打包 |
| **Checksums** | - | \`v2ray-subscription-manager-${VERSION}-checksums.txt\` | SHA256校验和 |

## 🔧 安装说明

### Windows
\`\`\`bash
# 下载并解压
unzip v2ray-subscription-manager-${VERSION}-windows-amd64.zip
# 直接运行
v2ray-subscription-manager-windows-amd64.exe --help
\`\`\`

### Linux/macOS
\`\`\`bash
# 下载并解压
unzip v2ray-subscription-manager-${VERSION}-linux-amd64.zip  # Linux
unzip v2ray-subscription-manager-${VERSION}-darwin-amd64.zip # macOS

# 添加执行权限
chmod +x v2ray-subscription-manager-*

# 运行
./v2ray-subscription-manager-linux-amd64 --help
\`\`\`

## 🔍 文件验证

使用SHA256校验和验证文件完整性：
\`\`\`bash
# 下载校验和文件
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/download/${VERSION}/v2ray-subscription-manager-${VERSION}-checksums.txt

# 验证文件
sha256sum -c v2ray-subscription-manager-${VERSION}-checksums.txt
\`\`\`

## 🚀 使用示例

\`\`\`bash
# 新的超时配置支持
./v2ray-subscription-manager start-proxy random https://your-subscription-url --timeout=30s

# 改进的测试功能
./v2ray-subscription-manager mvp-test https://your-subscription-url --show-config

# Windows平台优化体验
v2ray-subscription-manager.exe speed-test https://your-subscription-url --concurrency=50
\`\`\`

**完整文档**: [README.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/README.md)
**更新日志**: [RELEASE_NOTES.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/docs/RELEASE_NOTES.md)
EOF

echo "✅ 发布完成！"
echo ""
echo "📋 接下来的手动步骤："
echo "1. 访问: https://github.com/yxhpy/v2ray-subscription-manager/releases"
echo "2. 点击 'Create a new release'"
echo "3. 选择标签: $VERSION"
echo "4. 复制 release_notes_${VERSION}.md 的内容作为描述"
echo "5. 上传 bin/ 目录下的所有 *${VERSION}* 文件"
echo "6. 点击 'Publish release'"
echo ""
echo "📁 生成的文件位置: bin/"
echo "📄 Release说明文件: release_notes_${VERSION}.md" 