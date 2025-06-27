# v2.0.0 - 重大版本升级：增强清理系统、MVP测试器和双进程代理架构

## 🚀 重大特性升级

### 增强的清理系统
- **全新进程同步机制**: 实现更严格的进程停止和清理验证流程
- **智能清理逻辑**: 优化清理工具，增强进程同步机制和清理逻辑
- **重试机制**: 临时文件清理支持重试机制，确保清理结果的完整性和可靠性

### MVP测试器系统
- **新增MVP测试功能**: 添加MVP测试器支持，提供更专业的测试能力
- **测试工作流优化**: 改进测试流程，提高测试准确性和效率
- **状态管理增强**: 优化测试状态的跟踪和管理

### 双进程代理架构
- **双进程支持**: 支持更复杂的代理管理场景
- **进程间通信优化**: 优化进程间的协调和通信机制
- **资源隔离**: 提高不同代理进程间的资源隔离度

## 📦 下载文件

| 平台 | 架构 | 文件名 | 说明 |
|------|------|--------|------|
| **Windows** | x64 | `v2ray-subscription-manager-v2.0.0-windows-amd64.zip` | Windows 64位版本 |
| **Windows** | ARM64 | `v2ray-subscription-manager-v2.0.0-windows-arm64.zip` | Windows ARM64版本 |
| **Linux** | x64 | `v2ray-subscription-manager-v2.0.0-linux-amd64.zip` | Linux 64位版本 |
| **Linux** | ARM64 | `v2ray-subscription-manager-v2.0.0-linux-arm64.zip` | Linux ARM64版本 |
| **macOS** | x64 | `v2ray-subscription-manager-v2.0.0-darwin-amd64.zip` | macOS Intel版本 |
| **macOS** | ARM64 | `v2ray-subscription-manager-v2.0.0-darwin-arm64.zip` | macOS Apple Silicon版本 |
| **全平台** | 所有 | `v2ray-subscription-manager-v2.0.0-all-platforms.tar.gz` | 包含所有平台的压缩包 |
| **校验和** | - | `v2ray-subscription-manager-v2.0.0-checksums.txt` | SHA256校验和文件 |

## 🎯 升级指南

### Linux/macOS 用户
```bash
# 1. 下载对应平台的版本
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/download/v2.0.0/v2ray-subscription-manager-v2.0.0-linux-amd64.zip

# 2. 解压并安装
unzip v2ray-subscription-manager-v2.0.0-linux-amd64.zip
chmod +x v2ray-subscription-manager-linux-amd64
sudo mv v2ray-subscription-manager-linux-amd64 /usr/local/bin/v2ray-manager

# 3. 清理旧版本的临时文件（推荐）
v2ray-manager cleanup

# 4. 验证安装
v2ray-manager --version
```

### Windows 用户
```cmd
:: 1. 下载Windows版本
:: 下载 v2ray-subscription-manager-v2.0.0-windows-amd64.zip

:: 2. 解压到Program Files或其他目录
:: 解压 v2ray-subscription-manager-windows-amd64.exe

:: 3. 添加到PATH环境变量（可选）

:: 4. 验证安装
v2ray-subscription-manager-windows-amd64.exe --version
```

## 🔧 主要改进

### 系统清理功能
- ✅ 支持test_proxy_*文件清理
- ✅ 优化临时文件和状态文件的清理逻辑
- ✅ 提升资源管理效率，减少残留文件
- ✅ 严格的进程验证和同步机制

### 技术升级
- ✅ 新增fsnotify依赖，支持文件系统监控
- ✅ 更新系统依赖，提供更好的系统集成
- ✅ 优化编译参数，减小二进制文件大小
- ✅ 完善的自动化发布流程

### 文档和用户体验
- ✅ 更新README文档，增加系统清理工具的使用说明
- ✅ 提供详细的清理工具使用指南
- ✅ 添加新架构的使用说明和最佳实践

## 🐛 修复的问题
- ✅ 修复清理工具的进程同步问题
- ✅ 修复临时文件清理不完全的问题
- ✅ 修复MVP测试器的资源管理问题
- ✅ 修复双进程代理的状态同步问题
- ✅ 增强系统清理的可靠性和稳定性

## 📈 性能改进
- **清理效率**: 显著提升清理工具的执行效率
- **资源管理**: 更高效的资源管理和回收机制
- **系统响应**: 提高系统整体响应速度和稳定性

## ⚠️ 破坏性变更
- **架构升级**: 内部架构有重大改进，但保持API兼容性
- **清理行为**: 清理工具的行为更加严格和完善
- **配置变更**: 部分内部配置结构有调整

## 🔮 下个版本计划
- GUI界面开发
- 更多协议支持（VMess、Trojan）
- 性能监控和分析功能
- 配置文件管理增强

## 📊 文件大小对比

| 平台 | v1.3.0 | v2.0.0 | 变化 |
|------|--------|--------|------|
| Windows x64 | 2.6MB | 2.7MB | +3.8% |
| Linux x64 | 2.5MB | 2.6MB | +4.0% |
| macOS ARM64 | 2.4MB | 2.5MB | +4.2% |

*大小略有增加主要由于新增的MVP测试器和双进程架构功能*

---

**Full Changelog**: https://github.com/yxhpy/v2ray-subscription-manager/compare/v1.3.0...v2.0.0

感谢所有用户的支持和反馈！如有问题请在Issues中反映。 