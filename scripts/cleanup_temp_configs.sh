#!/bin/bash

echo "🧹 清理临时配置文件"
echo "=================="

# 清理V2Ray临时配置文件
echo "🔍 查找V2Ray临时配置文件..."
v2ray_configs=$(find . -name "temp_v2ray_config_*.json" -o -name "test_v2ray_config_*.json" 2>/dev/null | wc -l)
if [ $v2ray_configs -gt 0 ]; then
    echo "📁 找到 $v2ray_configs 个V2Ray临时配置文件"
    find . -name "temp_v2ray_config_*.json" -o -name "test_v2ray_config_*.json" 2>/dev/null | while read file; do
        echo "🗑️  删除: $file"
        rm -f "$file"
    done
else
    echo "✅ 没有找到V2Ray临时配置文件"
fi

# 清理Hysteria2临时配置文件
echo ""
echo "🔍 查找Hysteria2临时配置文件..."
if [ -d "./hysteria2" ]; then
    hysteria2_configs=$(find ./hysteria2 -name "config_*.yaml" -o -name "test_config_*.yaml" 2>/dev/null | wc -l)
    if [ $hysteria2_configs -gt 0 ]; then
        echo "📁 找到 $hysteria2_configs 个Hysteria2临时配置文件"
        find ./hysteria2 -name "config_*.yaml" -o -name "test_config_*.yaml" 2>/dev/null | while read file; do
            echo "🗑️  删除: $file"
            rm -f "$file"
        done
    else
        echo "✅ 没有找到Hysteria2临时配置文件"
    fi
else
    echo "✅ hysteria2目录不存在"
fi

# 清理其他可能的临时文件
echo ""
echo "🔍 查找其他临时文件..."
temp_files=$(find . -name "*.tmp" -o -name "proxy_state_*.json" 2>/dev/null | wc -l)
if [ $temp_files -gt 0 ]; then
    echo "📁 找到 $temp_files 个其他临时文件"
    find . -name "*.tmp" -o -name "proxy_state_*.json" 2>/dev/null | while read file; do
        echo "🗑️  删除: $file"
        rm -f "$file"
    done
else
    echo "✅ 没有找到其他临时文件"
fi

echo ""
echo "✨ 清理完成！" 