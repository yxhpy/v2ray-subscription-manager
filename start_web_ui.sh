#!/bin/bash

echo "🚀 启动 V2Ray 订阅管理器 Web UI"
echo "================================"

# 检查是否已经有进程在运行
if pgrep -f "web-ui" > /dev/null; then
    echo "⚠️  检测到 Web UI 已在运行，正在停止..."
    pkill -f "web-ui"
    sleep 2
fi

# 构建项目
echo "🔧 构建 Web UI..."
go build -o web-ui cmd/web-ui/main.go

if [ $? -ne 0 ]; then
    echo "❌ 构建失败！"
    exit 1
fi

echo "✅ 构建完成！"

# 启动服务器
echo "🌟 启动 Web UI 服务器..."
echo "📱 访问地址: http://localhost:8888"
echo "🛑 按 Ctrl+C 停止服务器"
echo "================================"

./web-ui