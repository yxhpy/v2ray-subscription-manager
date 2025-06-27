# v1.3.0 - 重大修复：解决SS订阅解析和测速工作流问题，成功率从0%提升到84%

## 🎯 重大修复和改进

### 🔧 核心问题修复
- **SS订阅解析问题修复**: 修复了base64解码问题，现在支持无padding的base64编码
- **测速工作流优化**: 重构了context管理逻辑，解决了`context canceled`错误
- **加密方法支持**: 新增对`chacha20-ietf-poly1305`等多种加密方法的支持
- **Windows兼容性**: 大幅改进Windows环境下的稳定性和性能

### 📈 性能提升
- **成功率提升**: 测速成功率从**0%提升到84.2%** (48/57个节点成功)
- **启动时间优化**: Windows环境下代理启动时间优化
- **网络连接**: 改进了HTTP客户端配置，提升连接稳定性
- **错误处理**: 更详细的错误信息和智能重试机制

### 🆕 新增功能
- **Windows优化文档**: 新增`docs/WINDOWS_OPTIMIZATION.md`优化指南
- **故障排除指南**: 新增`docs/WINDOWS_TROUBLESHOOTING.md`问题解决方案
- **智能重试**: 网络请求失败时自动重试最多3次
- **详细错误报告**: 更准确的错误分类和解决建议

### 🛠️ 技术改进
- **Base64解码**: 支持标准编码和无padding编码的自动切换
- **代理管理**: 优化了V2Ray和Hysteria2代理的启动检测
- **并发控制**: 降低默认并发数，避免系统资源耗尽
- **超时设置**: 增加超时时间，适应不同网络环境

## 📦 下载文件

| 平台 | 架构 | 文件名 | 说明 |
|------|------|--------|------|
| **Windows** | x64 | `v2ray-subscription-manager-v1.3.0-windows-amd64.zip` | Windows 64位版本 |
| **Windows** | ARM64 | `v2ray-subscription-manager-v1.3.0-windows-arm64.zip` | Windows ARM64版本 |
| **Linux** | x64 | `v2ray-subscription-manager-v1.3.0-linux-amd64.zip` | Linux 64位版本 |
| **Linux** | ARM64 | `v2ray-subscription-manager-v1.3.0-linux-arm64.zip` | Linux ARM64版本 |
| **macOS** | Intel | `v2ray-subscription-manager-v1.3.0-darwin-amd64.zip` | macOS Intel版本 |
| **macOS** | Apple Silicon | `v2ray-subscription-manager-v1.3.0-darwin-arm64.zip` | macOS M1/M2版本 |
| **All Platforms** | 通用 | `v2ray-subscription-manager-v1.3.0-all-platforms.tar.gz` | 所有平台打包 |
| **Checksums** | - | `v2ray-subscription-manager-v1.3.0-checksums.txt` | SHA256校验和 |

## 🔧 安装说明

### Windows
```bash
# 下载并解压
unzip v2ray-subscription-manager-v1.3.0-windows-amd64.zip
# 直接运行
v2ray-subscription-manager-windows-amd64.exe --help
```

### Linux/macOS
```bash
# 下载并解压
unzip v2ray-subscription-manager-v1.3.0-linux-amd64.zip  # Linux
unzip v2ray-subscription-manager-v1.3.0-darwin-amd64.zip # macOS

# 添加执行权限
chmod +x v2ray-subscription-manager-*

# 运行
./v2ray-subscription-manager-linux-amd64 --help
```

## 🔍 文件验证

使用SHA256校验和验证文件完整性：
```bash
# 下载校验和文件
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/download/v1.3.0/v2ray-subscription-manager-v1.3.0-checksums.txt

# 验证文件
sha256sum -c v2ray-subscription-manager-v1.3.0-checksums.txt
```

## 📊 构建信息

- **构建时间**: 2025-06-27 16:44:55
- **Go版本**: go1.24.4
- **Git提交**: a20ed8e
- **构建参数**: `-ldflags="-s -w"` (优化大小)

## 🚀 使用示例

```bash
# 测试订阅链接
./v2ray-subscription-manager parse https://your-subscription-url

# 启动代理
./v2ray-subscription-manager start-proxy random https://your-subscription-url

# 批量测速 (推荐配置)
./v2ray-subscription-manager speed-test-custom https://your-subscription-url --concurrency=5 --timeout=30

# Windows环境优化测速
./v2ray-subscription-manager speed-test-custom https://your-subscription-url --concurrency=2 --timeout=60
```

## 🆙 升级说明

如果你从旧版本升级，建议：

1. **删除旧版本**的二进制文件
2. **重新下载**适合你平台的新版本
3. **测试基本功能**：先用`parse`命令测试订阅解析
4. **查看新文档**：阅读Windows优化指南和故障排除文档

## 🐛 已知问题

- 部分老旧的SS节点可能仍然无法连接（约15%失败率）
- 极少数情况下Windows环境可能需要手动安装Visual C++运行库

**完整文档**: [README.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/README.md)
**更新日志**: [RELEASE_NOTES.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/RELEASE_NOTES.md)
**Windows优化**: [WINDOWS_OPTIMIZATION.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/docs/WINDOWS_OPTIMIZATION.md)
**故障排除**: [WINDOWS_TROUBLESHOOTING.md](https://github.com/yxhpy/v2ray-subscription-manager/blob/main/docs/WINDOWS_TROUBLESHOOTING.md)
