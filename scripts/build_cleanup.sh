#!/bin/bash

# 构建清理工具
echo "🔨 构建清理工具..."

# 确保脚本在项目根目录运行
cd "$(dirname "$0")/.." || exit 1

# 构建清理工具
echo "📦 编译cleanup工具..."
go build -o bin/cleanup cmd/cleanup.go

if [ $? -eq 0 ]; then
    echo "✅ 清理工具构建成功!"
    echo "📍 可执行文件位置: bin/cleanup"
    echo ""
    echo "使用方法："
    echo "  ./bin/cleanup    # 清理所有临时文件"
    echo ""
    
    # 设置执行权限
    chmod +x bin/cleanup
    echo "✅ 已设置执行权限"
else
    echo "❌ 构建失败!"
    exit 1
fi 