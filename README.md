# V2Ray Subscription Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/yxhpy/v2ray-subscription-manager)](https://goreportcard.com/report/github.com/yxhpy/v2ray-subscription-manager)
[![Release](https://img.shields.io/github/release/yxhpy/v2ray-subscription-manager.svg)](https://github.com/yxhpy/v2ray-subscription-manager/releases)

一个强大的V2Ray订阅链接解析器和代理管理器，支持多种协议解析、自动代理配置和智能节点管理。

## 🎯 项目亮点

- 🏗️ **模块化架构** - 采用清晰的分层架构设计，代码结构清晰，易于维护和扩展
- ⚡ **高性能并发** - 支持100+并发测速，优化的资源管理和智能端口分配
- 🔧 **自动化管理** - 一键安装依赖、自动配置、智能清理，使用简单
- 🌐 **多协议支持** - 支持VLESS、Shadowsocks、Hysteria2等主流协议
- 📊 **详细报告** - 生成完整的测速报告，支持JSON和文本格式导出
- 🔒 **稳定可靠** - 完善的错误处理和资源清理机制，确保系统稳定

## ✨ 特性

### 🔍 订阅解析
- ✅ 支持V2Ray订阅链接自动解析
- ✅ 支持多种协议：VLESS、Shadowsocks (SS)、Hysteria2
- ✅ 智能Base64解码和参数解析
- ✅ JSON格式结构化输出
- ✅ 错误节点自动过滤

### ⚡ 核心管理
- ✅ 自动下载V2Ray核心
- ✅ 自动下载Hysteria2客户端
- ✅ 跨平台支持（Windows、Linux、macOS）
- ✅ 多架构支持（amd64、arm64）
- ✅ 自动解压和权限设置
- ✅ 版本检查和更新提示

### 🚀 代理管理
- ✅ 一键启动随机节点代理
- ✅ 指定节点代理启动
- ✅ HTTP/SOCKS代理同时支持
- ✅ 智能端口分配，避免冲突
- ✅ 代理状态实时监控
- ✅ 连接测试和健康检查
- ✅ 状态持久化
- ✅ 优雅的进程管理

### 📊 测速工作流
- ✅ 批量节点测速（支持300+节点）
- ✅ 自定义测速参数
- ✅ 高并发测试优化（支持100+线程）
- ✅ 详细测速报告和统计信息
- ✅ 自动依赖检查和安装
- ✅ 智能资源管理和回收
- ✅ 进程和端口自动清理
- ✅ 支持多种测试目标（Google、百度等）
- ✅ 实时进度显示
- ✅ 测试结果排序和过滤

## 🏗️ 项目架构

```
verify_v2ray_ui2/
├── cmd/v2ray-manager/          # 主程序入口
├── internal/                   # 内部包
│   ├── core/                   # 核心功能模块
│   │   ├── downloader/         # 下载器模块
│   │   ├── parser/             # 订阅解析模块
│   │   ├── proxy/              # 代理管理模块
│   │   └── workflow/           # 工作流模块
│   ├── platform/               # 平台相关功能
│   └── utils/                  # 工具函数
├── pkg/types/                  # 公共类型定义
├── configs/                    # 配置文件模板
├── scripts/                    # 构建和发布脚本
└── docs/                       # 项目文档
```

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

# 构建主程序
go build -o v2ray-manager ./cmd/v2ray-manager/

# 或使用构建脚本
chmod +x scripts/build.sh
./scripts/build.sh
```

### 基本使用

#### 1. 准备工作
```bash
# 检查并下载V2Ray核心
./v2ray-manager check-v2ray
./v2ray-manager download-v2ray

# 检查并下载Hysteria2客户端（可选）
./v2ray-manager check-hysteria2
./v2ray-manager download-hysteria2

# 注意：测速工作流会自动检查和安装依赖，无需手动操作
```

#### 2. 解析订阅
```bash
# 解析订阅链接，查看所有节点
./v2ray-manager parse https://your-subscription-url

# 列出可用节点
./v2ray-manager list-nodes https://your-subscription-url
```

#### 3. 启动代理
```bash
# 随机启动一个节点
./v2ray-manager start-proxy random https://your-subscription-url

# 指定节点启动（节点索引从0开始）
./v2ray-manager start-proxy index https://your-subscription-url 5

# 启动Hysteria2代理
./v2ray-manager start-hysteria2 https://your-subscription-url 0
```

#### 4. 管理代理
```bash
# 查看代理状态
./v2ray-manager proxy-status

# 测试代理连接
./v2ray-manager test-proxy

# 停止代理
./v2ray-manager stop-proxy
```

#### 5. 测速工作流 ⭐
```bash
# 使用默认配置测速（自动检查依赖）
./v2ray-manager speed-test https://your-subscription-url

# 自定义测速参数 - 高性能测试
./v2ray-manager speed-test-custom https://your-subscription-url \
  --concurrency=100 \
  --timeout=30 \
  --test-url=https://www.google.com \
  --output=speed_test_results.txt

# 限制节点数量的快速测试
./v2ray-manager speed-test-custom https://your-subscription-url \
  --concurrency=50 \
  --timeout=20 \
  --max-nodes=50 \
  --test-url=http://www.baidu.com
```

#### 6. 系统清理 🧹
```bash
# 使用独立清理工具清理临时文件（推荐）
./bin/cleanup

# 手动构建清理工具
./scripts/build_cleanup.sh

# 清理所有临时文件和进程（如果主程序支持）
./v2ray-manager cleanup

# 强制终止所有相关进程
./v2ray-manager kill-all
```

**清理工具说明：**
- 自动清理 `auto_proxy_best_node.json`、`test_proxy_*.json` 等临时文件
- 清理所有代理配置文件和状态文件
- 安全清理，不会删除重要配置文件
- 详细清理日志，显示删除的文件
- 参见 [清理指南](docs/CLEANUP_GUIDE.md) 了解更多

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

### 清理命令
```bash
cleanup                             # 清理所有临时文件和进程
kill-all                            # 强制终止所有相关进程
```

**清理功能说明：**
- `cleanup`: 智能清理临时文件、终止相关进程、释放端口
- `kill-all`: 强制终止所有相关进程，包括V2Ray、Hysteria2、代理管理器等
- 独立脚本：`scripts/cleanup.sh` (Linux/macOS) 和 `scripts/cleanup.bat` (Windows)
- 支持跨平台清理，自动检测操作系统
- 保护用户配置文件，只清理临时文件

### 测试命令
```bash
test-hysteria2                      # 测试Hysteria2代理连接
```

### 测速工作流命令
```bash
speed-test <订阅链接>                # 默认测速工作流
speed-test-custom <订阅链接> [选项]   # 自定义测速工作流
```

**测速工作流选项：**
```bash
--concurrency=N          # 并发数量（默认：50）
--timeout=N              # 超时时间秒数（默认：30）
--test-url=URL           # 测试目标URL（默认：http://www.baidu.com）
--output=文件名           # 输出文件名（默认：speed_test_results.txt）
--max-nodes=N            # 最大测试节点数（默认：无限制）
```

## 📊 测速报告示例

测速完成后，程序会生成详细的测速报告：

```
V2Ray代理节点测速结果
测试时间: 2025-06-27 15:08:42
订阅链接: https://example.com/subscription
测试目标: http://www.baidu.com
总节点数: 298
================================================================================
成功节点: 91 个 (30.5%)
失败节点: 207 个 (69.5%)
平均延迟: 1686.9 ms
平均速度: 113.18 Mbps
最快节点: KZ_speednode_0049 (228.50 Mbps)
最慢节点: CY_speednode_0018 (7.19 Mbps)
================================================================================

📊 成功节点列表（按速度排序：快→慢）
排名 #1 - KZ_speednode_0049 (228.50 Mbps, 1269ms)
排名 #2 - US_speednode_0145 (189.80 Mbps, 921ms)
排名 #3 - US_speednode_0135 (188.41 Mbps, 1466ms)
...
```

## 🔧 配置

程序支持多种配置方式：

### 环境变量
```bash
export V2RAY_BIN_PATH="/usr/local/bin/v2ray"     # V2Ray二进制路径
export HYSTERIA2_BIN_PATH="/usr/local/bin/hysteria2"  # Hysteria2二进制路径
export CONFIG_DIR="./configs"                     # 配置文件目录
export LOG_LEVEL="info"                          # 日志级别
```

### 配置文件
程序会在以下位置查找配置文件：
- `./config.yaml`
- `~/.v2ray-manager/config.yaml`
- `/etc/v2ray-manager/config.yaml`

## 🔧 开发者发布

### 一键自动化发布（推荐）

**Linux/macOS:**
```bash
chmod +x scripts/release.sh
./scripts/release.sh v1.4.0 "项目重构完成，优化架构和性能"
```

**Windows:**
```cmd
scripts\release.bat v1.4.0 "项目重构完成，优化架构和性能"
```

**自动化发布脚本功能：**
- ✅ 版本验证和Git状态检查
- ✅ 跨平台编译（Windows/Linux/macOS × amd64/arm64）
- ✅ 自动打包和生成SHA256校验和
- ✅ Git提交、推送和标签创建
- ✅ 生成完整的GitHub Release说明文件

## 📝 更新日志

### v1.4.0 (2024-12-27)
- 🏗️ **重大重构**：采用模块化架构，提升代码可维护性
- ⚡ **性能优化**：优化并发处理和资源管理
- 🔧 **改进的工作流**：更稳定的测速工作流，支持大规模节点测试
- 📊 **增强报告**：更详细的测试报告和统计信息
- 🧹 **新增清理功能**：完善的系统清理机制，支持一键清理临时文件和进程
- 🔄 **自动代理服务**：新增高度可配置的自动代理服务，支持30+配置参数
- 🛠️ **Bug修复**：修复多个已知问题，提升稳定性
- 📚 **文档更新**：完善的项目文档和使用说明

### v1.3.0 (2024-12-20)
- ✅ 添加Hysteria2协议支持
- ✅ 重构测速工作流
- ✅ 优化并发处理
- ✅ 改进错误处理

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## ⭐ 项目支持

如果这个项目对您有帮助，请考虑给个⭐️星星支持！

## 📞 联系方式

- 项目地址：[GitHub](https://github.com/yxhpy/v2ray-subscription-manager)
- 问题反馈：[Issues](https://github.com/yxhpy/v2ray-subscription-manager/issues)

---

💡 **提示**：建议在生产环境使用前，先在测试环境中验证所有功能。 