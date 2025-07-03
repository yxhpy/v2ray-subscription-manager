#!/bin/bash

echo "🧪 V2Ray UI 全功能测试脚本"
echo "================================"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8888"

# 测试函数
test_api() {
    local name="$1"
    local url="$2"
    local method="${3:-GET}"
    local data="$4"
    
    echo -n "测试 $name... "
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" "$url")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ 成功${NC}"
        if echo "$body" | grep -q '"success":true'; then
            echo -e "   ${BLUE}API响应: 成功${NC}"
        else
            echo -e "   ${YELLOW}API响应: 可能有问题${NC}"
        fi
    else
        echo -e "${RED}❌ 失败 (HTTP $http_code)${NC}"
    fi
    
    echo
}

echo -e "${BLUE}1. 测试静态文件加载${NC}"
test_api "CSS样式文件" "$BASE_URL/static/css/style.css"
test_api "JavaScript文件" "$BASE_URL/static/js/app.js"

echo -e "${BLUE}2. 测试API接口${NC}"
test_api "系统状态API" "$BASE_URL/api/status"
test_api "订阅列表API" "$BASE_URL/api/subscriptions"

echo -e "${BLUE}3. 测试主页加载${NC}"
test_api "主页HTML" "$BASE_URL/"

echo -e "${BLUE}4. 测试添加订阅功能${NC}"
test_api "添加测试订阅" "$BASE_URL/api/subscriptions" "POST" '{
    "url": "https://example.com/test-subscription",
    "name": "自动测试订阅"
}'

echo -e "${BLUE}5. 验证订阅添加结果${NC}"
test_api "获取更新后的订阅列表" "$BASE_URL/api/subscriptions"

echo "================================"
echo -e "${GREEN}✨ 测试完成！${NC}"
echo
echo -e "${YELLOW}📱 访问地址:${NC}"
echo "   主界面: $BASE_URL"
echo "   测试页面: file://$(pwd)/test_frontend.html"
echo
echo -e "${YELLOW}🔧 如果有问题，请检查:${NC}"
echo "   1. Web UI服务器是否在运行 (端口8888)"
echo "   2. 静态文件路径是否正确"
echo "   3. 浏览器控制台是否有JavaScript错误"
echo "   4. CSS样式是否正确应用"