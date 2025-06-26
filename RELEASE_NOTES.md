# V2Ray Subscription Manager v1.0.0 Release Notes

## 🎉 首个正式版本发布

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