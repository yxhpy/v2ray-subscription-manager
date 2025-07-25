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

### 开发者发布

#### 一键自动化发布（推荐）

**Linux/macOS:**
```bash
# 给脚本添加执行权限
chmod +x release.sh

# 发布新版本
./release.sh v1.3.0 "添加新功能和修复Bug"
./release.sh v1.2.1 "修复Windows兼容性问题"
```

**Windows:**
```cmd
# 直接运行批处理脚本
release.bat v1.3.0 "添加新功能和修复Bug"
release.bat v1.2.1 "修复Windows兼容性问题"
```

**自动化发布脚本功能：**
- ✅ 版本验证和Git状态检查
- ✅ 跨平台编译（Windows/Linux/macOS × amd64/arm64）
- ✅ 自动打包和生成SHA256校验和
- ✅ Git提交、推送和标签创建
- ✅ 生成完整的GitHub Release说明文件

脚本执行完成后，只需手动在GitHub上创建Release并上传生成的文件。

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

## 🆕 最新改进 (v1.2.0)

### 🔧 Windows兼容性重大修复
- **Hysteria2并发下载修复**: 添加全局互斥锁防止多个worker同时下载Hysteria2，解决重复下载问题
- **Windows路径兼容性**: 自动处理Windows下的.exe扩展名，修复可执行文件路径问题
- **PE文件验证**: 添加Windows PE文件格式验证，确保下载的可执行文件有效
- **文件完整性检查**: 增强下载验证机制，包含文件大小和格式检查

### 🧠 智能解码系统
- **Base64智能检测**: 重写解码逻辑，能够智能识别内容是否为Base64编码
- **协议前缀检测**: 自动检测协议前缀（如`://`），避免误解码纯文本订阅内容
- **字符集验证**: 验证Base64字符集合法性，提高解析准确性

### 🧹 完善的临时文件清理
- **跨平台清理**: 针对Windows和Unix系统分别优化清理策略
- **唯一配置文件**: 为每个Hysteria2实例生成唯一配置文件路径（纳秒时间戳）
- **多重清理模式**: 支持多种文件名模式匹配，确保所有临时文件被清理
- **Windows文件锁处理**: 特殊处理Windows下的文件锁定问题，多次重试删除

### 🔄 下载重试机制
- **多次重试**: 下载失败时自动重试最多3次
- **多源支持**: 准备支持多个下载源（当前实现单源）
- **断点续传准备**: 为未来的断点续传功能预留接口

### 🚀 智能工作流系统
- **自动依赖检查**: 工作流启动时自动检查V2Ray和Hysteria2安装状态
- **一键安装**: 检测到缺失依赖时自动下载安装，无需手动操作
- **完整生命周期管理**: 依赖检查 → 自动安装 → 并发测试 → 资源回收 → 深度清理

### ⚡ 高性能测速引擎
- **超高并发**: 支持100+线程同时测试，充分利用系统性能
- **智能资源管理**: 每个测试完成后立即回收资源，避免资源堆积
- **进程和端口自动清理**: 测试完成后自动清理所有残留进程和端口占用
- **内存优化**: 优化内存使用，支持大规模节点测试

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