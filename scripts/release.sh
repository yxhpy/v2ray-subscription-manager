#!/bin/bash

# V2Ray Subscription Manager 自动化发布脚本
# 使用方法: ./release.sh <version> [commit_message]
# 示例: ./release.sh v1.2.0 "修复Windows兼容性问题"

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "\n${BLUE}🚀 $1${NC}"
}

# 检查参数
if [ $# -lt 1 ]; then
    print_error "使用方法: $0 <version> [commit_message]"
    print_info "示例: $0 v1.2.0 \"修复Windows兼容性问题\""
    exit 1
fi

VERSION=$1
COMMIT_MSG=${2:-"Release $VERSION"}

# 验证版本格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "版本格式错误，应该是 vX.Y.Z 格式，如 v1.2.0"
    exit 1
fi

print_step "开始自动化发布流程 - 版本: $VERSION"

# 1. 检查工作目录状态
print_step "检查Git工作目录状态"
if [ -n "$(git status --porcelain)" ]; then
    print_warning "工作目录有未提交的更改，将自动提交"
    git add .
    git status --short
else
    print_success "工作目录干净"
fi

# 2. 编译所有平台版本
print_step "编译所有平台版本"
mkdir -p bin

# 定义编译目标
declare -a targets=(
    "windows/amd64/.exe"
    "windows/arm64/.exe"
    "linux/amd64/"
    "linux/arm64/"
    "darwin/amd64/"
    "darwin/arm64/"
)

print_info "开始编译各平台版本..."

for target in "${targets[@]}"; do
    IFS='/' read -r goos goarch ext <<< "$target"
    output="bin/v2ray-subscription-manager-${goos}-${goarch}${ext}"
    
    print_info "编译 ${goos}/${goarch}..."
    GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w" -o "$output" .
    
    if [ -f "$output" ]; then
        size=$(ls -lh "$output" | awk '{print $5}')
        print_success "编译完成: $output ($size)"
    else
        print_error "编译失败: $output"
        exit 1
    fi
done

# 3. 创建压缩包
print_step "创建发布压缩包"
cd bin

# 清理旧版本压缩包
rm -f *${VERSION}*.zip *${VERSION}*.tar.gz

declare -a zip_files=()

for target in "${targets[@]}"; do
    IFS='/' read -r goos goarch ext <<< "$target"
    binary="v2ray-subscription-manager-${goos}-${goarch}${ext}"
    zipfile="v2ray-subscription-manager-${VERSION}-${goos}-${goarch}.zip"
    
    if [ -f "$binary" ]; then
        print_info "创建压缩包: $zipfile"
        zip -q "$zipfile" "$binary"
        zip_files+=("$zipfile")
        
        size=$(ls -lh "$zipfile" | awk '{print $5}')
        print_success "压缩包创建完成: $zipfile ($size)"
    fi
done

# 创建全平台压缩包
all_platforms_file="v2ray-subscription-manager-${VERSION}-all-platforms.tar.gz"
print_info "创建全平台压缩包: $all_platforms_file"
tar -czf "$all_platforms_file" v2ray-subscription-manager-*
size=$(ls -lh "$all_platforms_file" | awk '{print $5}')
print_success "全平台压缩包创建完成: $all_platforms_file ($size)"

cd ..

# 4. 生成校验和
print_step "生成文件校验和"
cd bin
sha256sum *${VERSION}* > "v2ray-subscription-manager-${VERSION}-checksums.txt"
print_success "校验和文件生成完成: v2ray-subscription-manager-${VERSION}-checksums.txt"
cd ..

# 5. 提交代码
print_step "提交代码到Git仓库"
git add .

# 生成详细的提交信息
cat > commit_message.tmp << EOF
release: $COMMIT_MSG

Version: $VERSION
Build Date: $(date '+%Y-%m-%d %H:%M:%S')
Platforms: Windows, Linux, macOS (amd64, arm64)

Changes in this release:
- 自动化构建和发布流程
- 跨平台二进制文件
- 完整的文件校验和
- 优化的构建参数 (-ldflags="-s -w")

Files included:
$(cd bin && ls -1 *${VERSION}* | sed 's/^/- /')
EOF

git commit -F commit_message.tmp
rm commit_message.tmp
print_success "代码提交完成"

# 6. 推送到远程仓库
print_step "推送代码到远程仓库"
git push origin main
print_success "代码推送完成"

# 7. 创建并推送标签
print_step "创建Git标签"

# 检查标签是否已存在
if git tag -l | grep -q "^${VERSION}$"; then
    print_warning "标签 $VERSION 已存在，删除旧标签"
    git tag -d "$VERSION"
    git push origin ":refs/tags/$VERSION" 2>/dev/null || true
fi

# 创建新标签
git tag -a "$VERSION" -m "Release $VERSION

Build Date: $(date '+%Y-%m-%d %H:%M:%S')
Commit: $(git rev-parse HEAD)

$COMMIT_MSG"

git push origin "$VERSION"
print_success "标签 $VERSION 创建并推送完成"

# 8. 生成Release说明
print_step "生成GitHub Release说明"
cat > "release_notes_${VERSION}.md" << EOF
# $VERSION - $(echo "$COMMIT_MSG" | sed 's/^Release [^ ]* - //')

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

## 📊 构建信息

- **构建时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **Go版本**: $(go version | awk '{print $3}')
- **Git提交**: $(git rev-parse --short HEAD)
- **构建参数**: \`-ldflags="-s -w"\` (优化大小)

## 🚀 使用示例

\`\`\`bash
# 测试订阅链接
./v2ray-subscription-manager parse https://your-subscription-url

# 启动代理
./v2ray-subscription-manager start-proxy random https://your-subscription-url

# 批量测速
./v2ray-subscription-manager speed-test https://your-subscription-url
\`\`\`

**完整文档**: [README.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/README.md)
**更新日志**: [RELEASE_NOTES.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/RELEASE_NOTES.md)
EOF

print_success "Release说明生成完成: release_notes_${VERSION}.md"

# 9. 显示发布摘要
print_step "发布摘要"
echo -e "\n${GREEN}🎉 发布完成！${NC}\n"

echo -e "${BLUE}📊 发布信息:${NC}"
echo -e "  版本: ${GREEN}$VERSION${NC}"
echo -e "  提交: ${GREEN}$(git rev-parse --short HEAD)${NC}"
echo -e "  时间: ${GREEN}$(date '+%Y-%m-%d %H:%M:%S')${NC}"

echo -e "\n${BLUE}📦 生成的文件:${NC}"
cd bin
for file in *${VERSION}*; do
    if [ -f "$file" ]; then
        size=$(ls -lh "$file" | awk '{print $5}')
        echo -e "  ${GREEN}✓${NC} $file (${size})"
    fi
done
cd ..

echo -e "\n${BLUE}🔗 下一步操作:${NC}"
echo -e "  1. 访问: ${YELLOW}https://github.com/yxhpy/v2ray-subscription-manager/releases${NC}"
echo -e "  2. 点击 '${YELLOW}Create a new release${NC}'"
echo -e "  3. 选择标签: ${YELLOW}$VERSION${NC}"
echo -e "  4. 复制 ${YELLOW}release_notes_${VERSION}.md${NC} 的内容作为描述"
echo -e "  5. 上传 ${YELLOW}bin/${NC} 目录下的所有 ${YELLOW}*${VERSION}*${NC} 文件"
echo -e "  6. 点击 '${YELLOW}Publish release${NC}'"

echo -e "\n${GREEN}✨ 自动化发布流程执行完成！${NC}" 