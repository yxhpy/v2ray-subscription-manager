# 双进程代理系统使用指南

## 概述

双进程代理系统是一个整合的解决方案，同时运行节点测试和代理服务两个进程，提供自动化的代理节点管理和服务。

## 系统架构

### 双进程设计
1. **节点测试进程**：定时获取订阅内容，测试所有节点性能，找出最佳节点并保存到状态文件
2. **代理服务进程**：监控状态文件变化，当发现新的最佳节点时自动切换代理服务

### 主要特性
- ✅ **自动化管理**：无需手动干预，系统自动维护最佳代理节点
- ✅ **热切换**：代理服务无缝切换到新的最佳节点
- ✅ **完整清理**：程序停止时自动清理所有临时文件和进程
- ✅ **多协议支持**：支持V2Ray和Hysteria2协议
- ✅ **性能优化**：智能测试算法，快速找出最佳节点

## 快速开始

### 1. 构建程序
```bash
go build -o dual-proxy ./cmd/dual-proxy
```

### 2. 基本使用
```bash
./dual-proxy -url "你的订阅链接"
```

### 3. 完整参数示例
```bash
./dual-proxy \
  -url "https://your-subscription-url" \
  -http-port 7890 \
  -socks-port 7891 \
  -interval 5m \
  -max-nodes 50 \
  -concurrency 5 \
  -state-file "best_node.json"
```

## 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-url` | string | 必需 | 订阅链接 |
| `-http-port` | int | 7890 | HTTP代理端口 |
| `-socks-port` | int | 7891 | SOCKS代理端口 |
| `-interval` | duration | 5m | 节点测试间隔 |
| `-max-nodes` | int | 50 | 最大测试节点数 |
| `-concurrency` | int | 5 | 测试并发数 |
| `-state-file` | string | mvp_best_node.json | 状态文件路径 |

## 使用示例

### 测试运行
使用提供的测试文件快速体验：
```bash
go run test_dual_proxy.go
```

### 生产环境运行
```bash
# 构建
go build -o dual-proxy ./cmd/dual-proxy

# 运行（替换为你的订阅链接）
./dual-proxy -url "https://your-subscription-url"
```

### 自定义配置运行
```bash
./dual-proxy \
  -url "https://your-subscription-url" \
  -http-port 8080 \
  -socks-port 8081 \
  -interval 3m \
  -max-nodes 30 \
  -concurrency 3
```

## 系统监控

### 日志输出
程序启动后会显示详细的状态信息：
- 🚀 系统启动状态
- 🧪 节点测试进度
- 🌐 代理服务状态
- 📊 测试结果摘要

### 状态文件
系统会生成状态文件（默认：`mvp_best_node.json`），包含：
- 当前最佳节点信息
- 测试统计数据
- 最后更新时间

## 停止服务

### 正常停止
按 `Ctrl+C` 停止程序，系统会：
1. 🛑 停止所有服务进程
2. 🧹 清理所有临时文件
3. 💀 终止相关代理进程
4. 🔌 清理端口占用

### 强制清理
如果程序异常退出，可以手动清理：
```bash
# 清理进程
pkill -f "v2ray|xray|hysteria"

# 清理临时文件
rm -f temp_*.json temp_*.yaml *.tmp *.temp
rm -f ./hysteria2/temp_*.yaml
```

## 代理配置

程序启动后，可以在浏览器或应用中配置代理：

### HTTP代理
- 地址：`127.0.0.1`
- 端口：`7890`（或自定义端口）

### SOCKS代理
- 地址：`127.0.0.1`
- 端口：`7891`（或自定义端口）

## 故障排除

### 常见问题

1. **依赖检查失败**
   - 确保系统中安装了v2ray或xray
   - 确保hysteria2可执行文件可用

2. **订阅解析失败**
   - 检查订阅链接是否有效
   - 检查网络连接

3. **端口占用**
   - 使用不同的端口参数
   - 检查是否有其他程序占用端口

4. **节点测试失败**
   - 检查防火墙设置
   - 尝试减少并发数

### 调试模式
程序会输出详细的日志信息，包括：
- 节点解析过程
- 测试结果详情
- 代理切换状态
- 错误信息

## 高级配置

### 性能调优
- 增加 `-concurrency` 参数提高测试速度
- 减少 `-interval` 参数更频繁地更新节点
- 调整 `-max-nodes` 参数控制测试范围

### 自定义测试URL
系统使用多个测试URL来验证节点性能：
- `http://www.google.com`
- `http://www.youtube.com`
- `http://www.facebook.com`

## 技术细节

### 文件监控
使用 `fsnotify` 库实现配置文件的实时监控和热重载。

### 进程管理
使用 Go 的 `context` 包实现优雅的进程生命周期管理。

### 资源清理
实现了完整的资源清理机制，包括：
- 临时配置文件清理
- 进程终止
- 端口占用清理

## 许可证

本项目遵循 MIT 许可证。 