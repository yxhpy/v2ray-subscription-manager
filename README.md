# V2Ray Subscription Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/yxhpy/v2ray-subscription-manager)](https://goreportcard.com/report/github.com/yxhpy/v2ray-subscription-manager)
[![Release](https://img.shields.io/github/release/yxhpy/v2ray-subscription-manager.svg)](https://github.com/yxhpy/v2ray-subscription-manager/releases)

一个强大的V2Ray订阅链接解析器和代理管理器，支持多种协议解析、自动代理配置和智能节点管理。

## ✨ 特性

### 🔍 订阅解析
- ✅ 支持V2Ray订阅链接自动解析
- ✅ 支持多种协议：VLESS、Shadowsocks (SS)、Hysteria2
- ✅ 智能Base64解码和参数解析
- ✅ JSON格式结构化输出

### ⚡ 核心管理
- ✅ 自动下载V2Ray核心
- ✅ 自动下载Hysteria2客户端
- ✅ 跨平台支持（Windows、Linux、macOS）
- ✅ 多架构支持（amd64、arm64）
- ✅ 自动解压和权限设置

### 🚀 代理管理
- ✅ 一键启动随机节点代理
- ✅ 指定节点代理启动
- ✅ HTTP/SOCKS代理同时支持
- ✅ 智能端口分配，避免冲突
- ✅ 代理状态实时监控
- ✅ 连接测试和健康检查
- ✅ 状态持久化

### 📊 测速工作流
- ✅ 批量节点测速
- ✅ 自定义测速参数
- ✅ 高并发测试优化（支持100+线程）
- ✅ 详细测速报告
- ✅ 自动依赖检查和安装
- ✅ 智能资源管理和回收
- ✅ 进程和端口自动清理
- ✅ 支持多种测试目标（Google、百度等）

## 📋 支持的协议

| 协议 | V2Ray支持 | Hysteria2支持 | 说明 |
|------|-----------|---------------|------|
| VLESS | ✅ | ❌ | 完全支持TLS、TCP等传输方式 |
| Shadowsocks | ✅ | ❌ | 自动转换加密方法兼容V2Ray 5.x |
| Hysteria2 | ❌ | ✅ | 使用独立Hysteria2客户端 |
| VMess | 🔄 | ❌ | 计划支持 |
| Trojan | 🔄 | ❌ | 计划支持 |

## 🚀 快速开始

### 安装

#### 方式1：下载预编译二进制文件
从 [Releases](https://github.com/yxhpy/v2ray-subscription-manager/releases) 页面下载适合您系统的版本。

#### 方式2：从源码构建
```bash
# 克隆仓库
git clone https://github.com/yxhpy/v2ray-subscription-manager.git
cd v2ray-subscription-manager

# 构建
./build.sh

# 或者简单构建
go build -o v2ray-subscription-manager .
```

### 基本使用

#### 1. 准备工作
```bash
# 检查并下载V2Ray核心
./v2ray-subscription-manager check-v2ray
./v2ray-subscription-manager download-v2ray

# 检查并下载Hysteria2客户端（可选）
./v2ray-subscription-manager check-hysteria2
./v2ray-subscription-manager download-hysteria2

# 注意：测速工作流会自动检查和安装依赖，无需手动操作
```

#### 2. 解析订阅
```bash
# 解析订阅链接，查看所有节点
./v2ray-subscription-manager parse https://your-subscription-url

# 列出可用节点
./v2ray-subscription-manager list-nodes https://your-subscription-url
```

#### 3. 启动代理
```bash
# 随机启动一个节点
./v2ray-subscription-manager start-proxy random https://your-subscription-url

# 指定节点启动（节点索引从0开始）
./v2ray-subscription-manager start-proxy index https://your-subscription-url 5

# 启动Hysteria2代理
./v2ray-subscription-manager start-hysteria2 https://your-subscription-url 0
```

#### 4. 管理代理
```bash
# 查看代理状态
./v2ray-subscription-manager proxy-status

# 测试代理连接
./v2ray-subscription-manager test-proxy

# 停止代理
./v2ray-subscription-manager stop-proxy
```

#### 5. 测速工作流
```bash
# 使用默认配置测速（自动检查依赖）
./v2ray-subscription-manager speed-test https://your-subscription-url

# 自定义测速参数 - 高性能测试
./v2ray-subscription-manager speed-test-custom https://your-subscription-url \
  --concurrency=100 \
  --timeout=30 \
  --test-url=https://www.google.com \
  --output=speed_test_results.txt

# 限制节点数量的快速测试
./v2ray-subscription-manager speed-test-custom https://your-subscription-url \
  --concurrency=50 \
  --timeout=20 \
  --max-nodes=50 \
  --test-url=http://www.baidu.com
```

## 📖 详细命令

### 订阅解析命令
```bash
parse <订阅链接>                    # 解析订阅链接
list-nodes <订阅链接>               # 列出所有可用节点
```

### 核心管理命令
```bash
download-v2ray                      # 下载V2Ray核心
check-v2ray                         # 检查V2Ray安装状态
download-hysteria2                  # 下载Hysteria2客户端
check-hysteria2                     # 检查Hysteria2安装状态
```

### 代理管理命令
```bash
start-proxy random <订阅链接>        # 随机启动代理
start-proxy index <订阅链接> <索引>  # 指定节点启动代理
start-hysteria2 <订阅链接> <索引>    # 启动Hysteria2代理
stop-proxy                          # 停止V2Ray代理
stop-hysteria2                      # 停止Hysteria2代理
proxy-status                        # 查看V2Ray代理状态
hysteria2-status                    # 查看Hysteria2代理状态
test-proxy                          # 测试V2Ray代理连接
test-hysteria2                      # 测试Hysteria2代理连接
```

### 测速工作流命令
```bash
speed-test <订阅链接>                # 默认配置测速（自动依赖检查）
speed-test-custom <订阅链接> [选项]   # 自定义测速（自动依赖检查）
  选项:
    --concurrency=数量              # 并发数（默认50，支持100+）
    --timeout=秒数                  # 超时时间（默认15秒）
    --output=文件名                 # 输出文件名
    --test-url=URL                  # 测试URL（默认百度）
    --max-nodes=数量                # 限制测试节点数量
```

## 📊 输出格式

### 解析结果
```json
{
  "total": 150,
  "nodes": [
    {
      "name": "🇺🇸 美国节点 | 4.4MB/s",
      "protocol": "vless",
      "server": "example.com",
      "port": "443",
      "uuid": "7f93e196-1b2f-4a42-8051-5815554c05db",
      "parameters": {
        "security": "tls",
        "sni": "example.com",
        "type": "tcp"
      }
    }
  ]
}
```

### 代理状态
```json
{
  "running": true,
  "http_port": 8080,
  "socks_port": 1080,
  "node_name": "🇺🇸 美国节点 | 4.4MB/s",
  "protocol": "vless",
  "server": "example.com"
}
```

## 🔧 高级特性

- **智能协议过滤**: 自动跳过不支持的协议，优先选择稳定节点
- **加密方法转换**: 自动将旧版SS加密方法转换为V2Ray 5.x兼容格式
- **端口智能分配**: 自动检测可用端口，避免冲突
- **状态持久化**: 程序重启后保持代理状态
- **健康检查**: 内置连接测试，确保代理可用性
- **并发优化**: 支持高并发测速和批量操作

## 🆕 最新改进 (v1.1.0)

### 🚀 智能工作流系统
- **自动依赖检查**: 工作流启动时自动检查V2Ray和Hysteria2安装状态
- **一键安装**: 检测到缺失依赖时自动下载安装，无需手动操作
- **完整生命周期管理**: 依赖检查 → 自动安装 → 并发测试 → 资源回收 → 深度清理

### ⚡ 高性能测速引擎
- **超高并发**: 支持100+线程同时测试，充分利用系统性能
- **智能资源管理**: 每个测试完成后立即回收资源，避免资源堆积
- **进程和端口自动清理**: 测试完成后自动清理所有残留进程和端口占用
- **内存优化**: 优化内存使用，支持大规模节点测试

### 🧹 完善的资源回收机制
- **实时清理**: 每个节点测试完成后立即清理相关资源
- **深度清理**: 工作流结束时执行深度清理，确保无残留
- **智能端口管理**: 为每个worker分配独立端口范围，避免冲突
- **临时文件清理**: 自动清理所有临时配置文件

### 📊 增强的测试功能
- **灵活测试目标**: 支持自定义测试URL（Google、百度等）
- **节点数量限制**: 支持限制测试节点数量，快速测试
- **详细进度显示**: 实时显示测试进度和当前节点信息
- **多格式输出**: 同时生成TXT和JSON格式的详细报告

### 🔧 使用示例
```bash
# 高性能Google连通性测试（100线程）
./v2ray-subscription-manager speed-test-custom https://your-subscription-url \
  --concurrency=100 \
  --timeout=30 \
  --test-url=https://www.google.com \
  --output=google_test_results.txt

# 快速测试前50个节点（50线程）
./v2ray-subscription-manager speed-test-custom https://your-subscription-url \
  --concurrency=50 \
  --max-nodes=50 \
  --test-url=http://www.baidu.com
```

## 📁 项目结构

```
v2ray-subscription-manager/
├── main.go                 # 主程序入口
├── parser.go               # 订阅解析模块
├── proxy.go                # V2Ray代理管理
├── hysteria2_proxy.go      # Hysteria2代理管理
├── downloader.go           # V2Ray下载器
├── hysteria2_downloader.go # Hysteria2下载器
├── workflow.go             # 测速工作流
├── build.sh                # 构建脚本
├── go.mod                  # Go模块文件
├── LICENSE                 # MIT许可证
├── README.md               # 项目文档
├── .gitignore              # Git忽略文件
├── v2ray/                  # V2Ray核心文件
├── hysteria2/              # Hysteria2文件
└── bin/                    # 构建输出目录
```

## 🛠️ 开发

### 环境要求
- Go 1.21+
- Git

### 构建
```bash
# 开发构建
go build -o v2ray-subscription-manager .

# 发布构建（所有平台）
./build.sh v1.0.0

# 运行测试
go test ./...
```

### 贡献指南
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## ⚠️ 免责声明

本工具仅供学习和研究使用，请遵守当地法律法规。使用者应对使用本工具产生的任何后果负责。

## 🙏 致谢

- [V2Ray](https://github.com/v2fly/v2ray-core) - 强大的网络代理工具
- [Hysteria2](https://github.com/apernet/hysteria) - 高性能代理协议
- 所有贡献者和用户的支持

---

**如果这个项目对您有帮助，请给个 ⭐ Star！** 