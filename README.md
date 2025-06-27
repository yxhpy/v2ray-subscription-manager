<div align="center">

# 🚀 V2Ray 订阅管理器

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/yxhpy/v2ray-subscription-manager)](https://goreportcard.com/report/github.com/yxhpy/v2ray-subscription-manager)
[![Release](https://img.shields.io/github/release/yxhpy/v2ray-subscription-manager.svg)](https://github.com/yxhpy/v2ray-subscription-manager/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-green.svg)](https://github.com/yxhpy/v2ray-subscription-manager)

**🌟 一个现代化、高性能的 V2Ray 订阅管理和代理测速工具**

*支持多协议解析 | 智能代理管理 | 高并发测速 | 双进程架构*

</div>

---

## 📖 项目简介

V2Ray 订阅管理器是一个功能强大的代理管理工具，专为现代网络环境设计。它不仅能够解析和管理 V2Ray 订阅链接，还提供了智能的代理切换、高性能测速和自动化管理功能。

### 🎯 核心亮点

- 🏗️ **模块化架构** - 采用清晰的分层设计，代码结构优雅，易于维护和扩展
- ⚡ **极致性能** - 支持 100+ 并发测速，智能资源管理和端口分配
- 🤖 **智能自动化** - 一键安装依赖、自动配置、智能清理，使用简单
- 🌐 **多协议支持** - 完整支持 VLESS、Shadowsocks、Hysteria2 等主流协议
- 📊 **专业报告** - 生成详细的测速报告，支持多种格式导出
- 🔒 **稳定可靠** - 完善的错误处理和资源清理机制，确保系统稳定运行

---

## ✨ 功能特性

<table>
<tr>
<td width="50%">

### 🔍 订阅解析
- ✅ V2Ray 订阅链接自动解析
- ✅ 多协议支持：VLESS、SS、Hysteria2
- ✅ 智能 Base64 解码和参数解析
- ✅ JSON 格式结构化输出
- ✅ 错误节点自动过滤

### ⚡ 核心管理
- ✅ 自动下载 V2Ray 核心
- ✅ 自动下载 Hysteria2 客户端
- ✅ 跨平台支持（Windows、Linux、macOS）
- ✅ 多架构支持（amd64、arm64）
- ✅ 自动解压和权限设置
- ✅ 版本检查和更新提示

</td>
<td width="50%">

### 🚀 代理管理
- ✅ 一键启动随机/指定节点代理
- ✅ HTTP/SOCKS 代理同时支持
- ✅ 智能端口分配，避免冲突
- ✅ 代理状态实时监控
- ✅ 连接测试和健康检查
- ✅ 状态持久化和恢复

### 📊 测速工作流
- ✅ 批量节点测速（支持 300+ 节点）
- ✅ 高并发测试优化（100+ 线程）
- ✅ 详细测速报告和统计
- ✅ 自动依赖检查和安装
- ✅ 智能资源管理和回收
- ✅ 实时进度显示

</td>
</tr>
</table>

---

## 🏗️ 项目架构

```
verify_v2ray_ui2/
├── 📁 cmd/                         # 程序入口
│   ├── v2ray-manager/              # 主程序
│   └── cleanup.go                  # 清理工具
├── 📁 internal/                    # 内部核心模块
│   ├── core/                       # 核心功能
│   │   ├── downloader/             # 📥 下载器模块
│   │   ├── parser/                 # 🔍 订阅解析模块
│   │   ├── proxy/                  # 🌐 代理管理模块
│   │   └── workflow/               # ⚡ 工作流模块
│   ├── platform/                   # 🖥️ 平台相关功能
│   └── utils/                      # 🔧 工具函数
├── 📁 pkg/types/                   # 📋 公共类型定义
├── 📁 configs/                     # ⚙️ 配置文件模板
├── 📁 scripts/                     # 📜 构建和发布脚本
└── 📁 docs/                        # 📚 项目文档
```

---

## 📋 协议支持

<div align="center">

| 协议 | V2Ray 支持 | Hysteria2 支持 | 状态 | 说明 |
|:----:|:----------:|:--------------:|:----:|:-----|
| **VLESS** | ✅ | ❌ | 🟢 完整支持 | 完全支持 TLS、TCP 等传输方式 |
| **Shadowsocks** | ✅ | ❌ | 🟢 完整支持 | 自动转换加密方法兼容 V2Ray 5.x |
| **Hysteria2** | ❌ | ✅ | 🟢 完整支持 | 使用独立 Hysteria2 客户端 |
| **VMess** | 🔄 | ❌ | 🟡 计划支持 | 下一版本将支持 |
| **Trojan** | 🔄 | ❌ | 🟡 计划支持 | 下一版本将支持 |

</div>

---

## 🚀 快速开始

### 📦 安装方式

<details>
<summary><b>方式一：下载预编译二进制文件（推荐）</b></summary>

从 [Releases](https://github.com/yxhpy/v2ray-subscription-manager/releases) 页面下载适合您系统的版本：

```bash
# Linux amd64
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/latest/download/v2ray-subscription-manager-linux-amd64
chmod +x v2ray-subscription-manager-linux-amd64
mv v2ray-subscription-manager-linux-amd64 v2ray-manager

# macOS amd64
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/latest/download/v2ray-subscription-manager-darwin-amd64
chmod +x v2ray-subscription-manager-darwin-amd64
mv v2ray-subscription-manager-darwin-amd64 v2ray-manager

# Windows amd64
# 下载 v2ray-subscription-manager-windows-amd64.exe
```

</details>

<details>
<summary><b>方式二：从源码构建</b></summary>

```bash
# 克隆仓库
git clone https://github.com/yxhpy/v2ray-subscription-manager.git
cd v2ray-subscription-manager

# 使用构建脚本（推荐）
chmod +x scripts/build.sh
./scripts/build.sh

# 或手动构建
go build -o v2ray-manager ./cmd/v2ray-manager/

# 构建清理工具
go build -o bin/cleanup ./cmd/cleanup.go
```

</details>

### 🎮 基本使用

#### 1️⃣ 准备工作

```bash
# 🔧 自动检查和下载依赖（推荐）
./v2ray-manager check-v2ray && ./v2ray-manager download-v2ray
./v2ray-manager check-hysteria2 && ./v2ray-manager download-hysteria2

# 💡 提示：测速工作流会自动检查和安装依赖，无需手动操作
```

#### 2️⃣ 解析订阅

```bash
# 📥 解析订阅链接，查看所有节点
./v2ray-manager parse https://your-subscription-url

# 📋 列出可用节点（带索引）
./v2ray-manager list-nodes https://your-subscription-url
```

#### 3️⃣ 启动代理

```bash
# 🎲 随机启动一个节点
./v2ray-manager start-proxy random https://your-subscription-url

# 🎯 指定节点启动（节点索引从0开始）
./v2ray-manager start-proxy index https://your-subscription-url 5

# ⚡ 启动 Hysteria2 代理
./v2ray-manager start-hysteria2 https://your-subscription-url 0
```

#### 4️⃣ 管理代理

```bash
# 📊 查看代理状态
./v2ray-manager proxy-status

# 🔍 测试代理连接
./v2ray-manager test-proxy

# 🛑 停止代理
./v2ray-manager stop-proxy
```

---

## ⭐ 高级功能

### 🏃‍♂️ 测速工作流

<details>
<summary><b>📊 基础测速</b></summary>

```bash
# 使用默认配置测速（自动检查依赖）
./v2ray-manager speed-test https://your-subscription-url
```

</details>

<details>
<summary><b>⚡ 高性能自定义测速</b></summary>

```bash
# 高性能测试配置
./v2ray-manager speed-test-custom https://your-subscription-url \
  --concurrency=100 \
  --timeout=30 \
  --test-url=https://www.google.com \
  --output=speed_test_results.txt

# 快速测试配置
./v2ray-manager speed-test-custom https://your-subscription-url \
  --concurrency=50 \
  --timeout=20 \
  --max-nodes=50 \
  --test-url=http://www.baidu.com
```

</details>

### 🤖 自动代理管理

<details>
<summary><b>🔄 智能自动代理</b></summary>

```bash
# 启动自动代理管理器（智能切换最佳节点）
./v2ray-manager auto-proxy https://your-subscription-url \
  --http-port=7890 \
  --socks-port=7891 \
  --interval=10 \
  --concurrency=20 \
  --max-nodes=100

# 高频更新配置
./v2ray-manager auto-proxy https://your-subscription-url \
  --interval=5 \
  --concurrency=30 \
  --timeout=20 \
  --test-url=https://www.google.com
```

</details>

### 🚀 MVP 双进程模式

<details>
<summary><b>⚡ 轻量级双进程方案</b></summary>

```bash
# 方案一：一键启动双进程系统
./v2ray-manager dual-proxy https://your-subscription-url \
  --http-port=8080 \
  --socks-port=1080

# 方案二：分别启动进程
# 进程1：启动 MVP 测试器
./v2ray-manager mvp-tester https://your-subscription-url \
  --interval=10 \
  --max-nodes=30 \
  --state-file=mvp_best_node.json

# 进程2：启动代理服务器
./v2ray-manager proxy-server mvp_best_node.json \
  --http-port=8080 \
  --socks-port=1080
```

</details>

### 🧹 系统清理

<details>
<summary><b>🔧 清理工具</b></summary>

```bash
# 使用独立清理工具（推荐）
./bin/cleanup

# 手动构建清理工具
./scripts/build_cleanup.sh && ./bin/cleanup

# 清理功能说明：
# ✅ 清理所有临时文件和状态文件
# ✅ 终止所有相关进程
# ✅ 安全清理，不删除重要配置
# ✅ 详细清理日志显示
```

</details>

---

## 📖 命令参考

<details>
<summary><b>📥 订阅解析命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `parse <订阅链接>` | 解析订阅链接 | `parse https://example.com/sub` |
| `list-nodes <订阅链接>` | 列出所有可用节点 | `list-nodes https://example.com/sub` |

</details>

<details>
<summary><b>⚙️ 核心管理命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `download-v2ray` | 下载 V2Ray 核心 | `download-v2ray` |
| `check-v2ray` | 检查 V2Ray 安装状态 | `check-v2ray` |
| `download-hysteria2` | 下载 Hysteria2 客户端 | `download-hysteria2` |
| `check-hysteria2` | 检查 Hysteria2 安装状态 | `check-hysteria2` |

</details>

<details>
<summary><b>🌐 代理管理命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `start-proxy random <订阅链接>` | 随机启动代理 | `start-proxy random https://example.com/sub` |
| `start-proxy index <订阅链接> <索引>` | 指定节点启动代理 | `start-proxy index https://example.com/sub 5` |
| `start-hysteria2 <订阅链接> <索引>` | 启动 Hysteria2 代理 | `start-hysteria2 https://example.com/sub 0` |
| `stop-proxy` | 停止 V2Ray 代理 | `stop-proxy` |
| `stop-hysteria2` | 停止 Hysteria2 代理 | `stop-hysteria2` |
| `proxy-status` | 查看 V2Ray 代理状态 | `proxy-status` |
| `hysteria2-status` | 查看 Hysteria2 代理状态 | `hysteria2-status` |
| `test-proxy` | 测试 V2Ray 代理连接 | `test-proxy` |
| `test-hysteria2` | 测试 Hysteria2 代理连接 | `test-hysteria2` |

</details>

<details>
<summary><b>📊 测速工作流命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `speed-test <订阅链接>` | 默认配置测速 | `speed-test https://example.com/sub` |
| `speed-test-custom <订阅链接> [选项]` | 自定义测速 | `speed-test-custom https://example.com/sub --concurrency=100` |

**自定义选项：**
- `--concurrency=数量` - 并发数（默认：50）
- `--timeout=秒数` - 超时时间（默认：30）
- `--output=文件名` - 输出文件（默认：speed_test_results.txt）
- `--test-url=URL` - 测试 URL（默认：http://www.google.com）
- `--max-nodes=数量` - 最大测试节点数（默认：无限制）

</details>

<details>
<summary><b>🤖 自动代理管理命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `auto-proxy <订阅链接> [选项]` | 启动自动代理管理器 | `auto-proxy https://example.com/sub --http-port=7890` |

**自动代理选项：**
- `--http-port=端口` - HTTP 代理端口（默认：7890）
- `--socks-port=端口` - SOCKS 代理端口（默认：7891）
- `--interval=分钟` - 更新间隔分钟数（默认：10）
- `--concurrency=数量` - 测试并发数（默认：20）
- `--timeout=秒数` - 测试超时秒数（默认：30）
- `--test-url=URL` - 测试 URL（默认：http://www.google.com）
- `--max-nodes=数量` - 最大测试节点数（默认：100）
- `--min-nodes=数量` - 最少通过节点数（默认：5）
- `--state-file=路径` - 状态文件路径
- `--valid-file=路径` - 有效节点文件路径
- `--no-auto-switch` - 禁用自动切换

</details>

<details>
<summary><b>🚀 MVP 双进程命令</b></summary>

| 命令 | 说明 | 示例 |
|------|------|------|
| `mvp-tester <订阅链接> [选项]` | 启动 MVP 节点测试器 | `mvp-tester https://example.com/sub --interval=10` |
| `proxy-server <配置文件> [选项]` | 启动代理服务器 | `proxy-server mvp_best_node.json --http-port=8080` |
| `dual-proxy <订阅链接> [选项]` | 启动双进程代理系统 | `dual-proxy https://example.com/sub --http-port=8080` |

</details>

---

## 🛠️ 开发指南

### 📋 系统要求

- **Go 版本**: 1.21+
- **操作系统**: Windows, Linux, macOS
- **架构**: amd64, arm64

### 🏗️ 构建项目

```bash
# 克隆项目
git clone https://github.com/yxhpy/v2ray-subscription-manager.git
cd v2ray-subscription-manager

# 安装依赖
go mod tidy

# 构建所有平台版本
./scripts/build.sh

# 构建当前平台版本
go build -o v2ray-manager ./cmd/v2ray-manager/
```

### 🚀 发布流程

```bash
# Linux/macOS 自动化发布
chmod +x scripts/release.sh
./scripts/release.sh v1.3.0 "添加新功能和修复Bug"

# Windows 自动化发布
scripts/release.bat v1.3.0 "添加新功能和修复Bug"
```

---

## 📚 文档资源

<div align="center">

| 📖 文档 | 🔗 链接 | 📝 说明 |
|:-------:|:-------:|:-------|
| **清理指南** | [CLEANUP_GUIDE.md](docs/CLEANUP_GUIDE.md) | 系统清理详细说明 |
| **双进程指南** | [DUAL_PROXY_GUIDE.md](docs/DUAL_PROXY_GUIDE.md) | 双进程架构使用指南 |
| **Windows 优化** | [WINDOWS_OPTIMIZATION.md](docs/WINDOWS_OPTIMIZATION.md) | Windows 系统优化建议 |
| **故障排除** | [WINDOWS_TROUBLESHOOTING.md](docs/WINDOWS_TROUBLESHOOTING.md) | 常见问题解决方案 |
| **发布说明** | [RELEASE_NOTES.md](docs/RELEASE_NOTES.md) | 版本更新日志 |

</div>

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 🐛 报告问题
- 使用 [Issues](https://github.com/yxhpy/v2ray-subscription-manager/issues) 报告 Bug
- 提供详细的错误信息和复现步骤

### 💡 功能建议
- 在 [Issues](https://github.com/yxhpy/v2ray-subscription-manager/issues) 中提出新功能建议
- 详细描述功能需求和使用场景

### 🔧 代码贡献
1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

---

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源协议。

---

## 🙏 致谢

感谢以下开源项目：

- [V2Ray](https://github.com/v2fly/v2ray-core) - 强大的代理工具
- [Hysteria](https://github.com/apernet/hysteria) - 高性能代理协议
- [fsnotify](https://github.com/fsnotify/fsnotify) - 文件系统监控

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！**

[![GitHub stars](https://img.shields.io/github/stars/yxhpy/v2ray-subscription-manager.svg?style=social&label=Star)](https://github.com/yxhpy/v2ray-subscription-manager)
[![GitHub forks](https://img.shields.io/github/forks/yxhpy/v2ray-subscription-manager.svg?style=social&label=Fork)](https://github.com/yxhpy/v2ray-subscription-manager)

**📧 联系我们**: [Issues](https://github.com/yxhpy/v2ray-subscription-manager/issues) | [Discussions](https://github.com/yxhpy/v2ray-subscription-manager/discussions)

</div> 