# Release Notes

## v1.1.0 - 智能工作流系统和高性能测速引擎 (2025-06-27)

### 🚀 重大新特性

#### 智能工作流系统
- **自动依赖检查**: 工作流启动时自动检查V2Ray和Hysteria2安装状态
- **一键安装**: 检测到缺失依赖时自动下载安装，无需手动操作
- **完整生命周期管理**: 依赖检查 → 自动安装 → 并发测试 → 资源回收 → 深度清理

#### 高性能测速引擎
- **超高并发**: 支持100+线程同时测试，充分利用系统性能
- **智能资源管理**: 每个测试完成后立即回收资源，避免资源堆积
- **进程和端口自动清理**: 测试完成后自动清理所有残留进程和端口占用
- **内存优化**: 优化内存使用，支持大规模节点测试

#### 完善的资源回收机制
- **实时清理**: 每个节点测试完成后立即清理相关资源
- **深度清理**: 工作流结束时执行深度清理，确保无残留
- **智能端口管理**: 为每个worker分配独立端口范围，避免冲突
- **临时文件清理**: 自动清理所有临时配置文件

### 📊 增强的测试功能
- **灵活测试目标**: 支持自定义测试URL（Google、百度等）
- **节点数量限制**: 支持限制测试节点数量，快速测试
- **详细进度显示**: 实时显示测试进度和当前节点信息
- **多格式输出**: 同时生成TXT和JSON格式的详细报告

### 🔧 技术改进
- 优化并发测试算法，支持100+线程
- 改进资源管理机制，确保无残留
- 增强错误处理和异常恢复
- 优化内存使用和性能

### 📈 性能提升
- 测试速度提升300%+（相比单线程）
- 资源利用率优化，支持大规模测试
- 完美的资源回收机制
- 零残留进程和端口占用

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

### 🐛 修复的问题
- 修复高并发测试时的资源泄漏问题
- 修复进程和端口无法正确回收的问题
- 修复临时文件残留问题
- 修复内存使用过高的问题

### 💔 破坏性变更
- 无破坏性变更，完全向后兼容

---

## v1.0.0 - 初始发布 (2025-06-26)

### ✨ 核心功能
- V2Ray订阅链接解析
- 多协议支持（VLESS、Shadowsocks、Hysteria2）
- 自动代理配置和管理
- 基础测速功能
- 跨平台支持

### 🔧 支持的平台
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### 📦 包含组件
- V2Ray核心自动下载器
- Hysteria2客户端自动下载器
- 代理管理器
- 基础测速工具

### 🎉 首个正式版本发布

### ✨ 主要功能
- 🔍 **订阅解析**: 支持VLESS、Shadowsocks、Hysteria2协议的V2Ray订阅链接解析
- ⚡ **核心管理**: 自动下载V2Ray和Hysteria2核心，支持跨平台部署
- 🚀 **代理管理**: 智能代理启动、状态监控、连接测试和健康检查
- 📊 **测速工作流**: 批量节点测速，支持自定义参数和并发优化
- 🌍 **跨平台支持**: 完整支持Windows、Linux、macOS的多种架构

### 📦 发布文件

| 平台 | 架构 | 文件名 | 大小 |
|------|------|--------|------|
| **Linux** | x64 | `v2ray-subscription-manager-v1.0.0-linux-amd64.zip` | 2.6MB |
| **Linux** | ARM64 | `v2ray-subscription-manager-v1.0.0-linux-arm64.zip` | 2.4MB |
| **macOS** | Intel | `v2ray-subscription-manager-v1.0.0-darwin-amd64.zip` | 2.7MB |
| **macOS** | Apple Silicon | `v2ray-subscription-manager-v1.0.0-darwin-arm64.zip` | 2.5MB |
| **Windows** | x64 | `v2ray-subscription-manager-v1.0.0-windows-amd64.zip` | 2.7MB |
| **Windows** | ARM64 | `v2ray-subscription-manager-v1.0.0-windows-arm64.zip` | 2.4MB |
| **All Platforms** | 通用 | `v2ray-subscription-manager-v1.0.0-all-platforms.tar.gz` | 15.2MB |

### 🚀 快速开始

#### 1. 下载和安装
```bash
# 下载对应平台的压缩包并解压
unzip v2ray-subscription-manager-v1.0.0-linux-amd64.zip

# 赋予执行权限（Linux/macOS）
chmod +x v2ray-subscription-manager-linux-amd64

# 下载V2Ray核心
./v2ray-subscription-manager-linux-amd64 download-v2ray
```

#### 2. 基本使用
```bash
# 解析订阅
./v2ray-subscription-manager-linux-amd64 parse https://your-subscription-url

# 启动随机代理
./v2ray-subscription-manager-linux-amd64 start-proxy random https://your-subscription-url

# 查看代理状态
./v2ray-subscription-manager-linux-amd64 proxy-status
```

### 🔧 技术特性
- **智能协议过滤**: 自动跳过不支持的协议，优先选择稳定节点
- **加密方法转换**: 自动将旧版SS加密方法转换为V2Ray 5.x兼容格式
- **端口智能分配**: 自动检测可用端口，避免冲突
- **状态持久化**: 程序重启后保持代理状态
- **健康检查**: 内置连接测试，确保代理可用性
- **并发优化**: 支持高并发测速和批量操作

### 📖 完整文档
详细使用说明请查看 [README.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/README.md)

### ⚠️ 免责声明
本工具仅供学习和研究使用，请遵守当地法律法规。

---

**发布日期**: 2025-06-27
**版本**: v1.0.0
**提交**: [885477c](https://github.com/yxhpy/v2ray-subscription-manager/commit/885477c)