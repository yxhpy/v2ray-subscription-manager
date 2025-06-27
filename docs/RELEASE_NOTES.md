# Release Notes

## v2.1.0 - 代理管理器和MVP测试器优化版本 - 增强超时处理、配置展示和Windows平台并发控制 (2025-01-28)

### 🚀 核心功能优化

#### 代理管理器增强
- **超时处理优化**: 改进代理连接的超时机制，提供更灵活的超时配置
- **配置展示功能**: 新增详细的代理配置展示，方便用户了解当前代理设置
- **状态监控**: 增强代理状态的实时监控和反馈机制
- **连接稳定性**: 优化代理连接的稳定性和可靠性

#### MVP测试器系统升级
- **测试URL配置**: 支持用户自定义测试URL，提供更灵活的测试选项
- **超时时间配置**: 允许用户配置测试超时时间，适应不同网络环境
- **测试结果优化**: 改进测试结果的展示和分析功能
- **错误处理增强**: 更完善的错误处理和异常恢复机制

#### Windows平台专项优化
- **并发控制**: 针对Windows平台优化并发控制机制，提高稳定性
- **进程管理**: 改进Windows下的进程管理和资源清理
- **兼容性增强**: 提升Windows平台的整体兼容性和性能
- **错误处理**: 优化Windows特有的错误处理逻辑

### 🔧 技术改进

#### 用户体验优化
- **配置可视化**: 新增配置信息的可视化展示功能
- **实时反馈**: 提供更详细的操作反馈和状态信息
- **错误提示**: 改进错误提示的准确性和友好性
- **操作指导**: 增强用户操作的指导和帮助信息

#### 系统稳定性
- **资源管理**: 优化系统资源的使用和管理
- **内存优化**: 改进内存使用效率，减少内存泄漏
- **进程同步**: 增强进程间的同步和协调机制
- **异常恢复**: 提升系统的异常恢复能力

### 🐛 修复的问题

#### 高优先级修复
- ✅ 修复代理管理器的超时处理问题
- ✅ 修复MVP测试器的配置展示异常
- ✅ 修复Windows平台的并发控制问题
- ✅ 修复用户配置的保存和加载问题

#### 稳定性修复
- ✅ 修复长时间运行时的内存泄漏问题
- ✅ 修复进程间通信的同步问题
- ✅ 修复Windows下的文件锁定问题
- ✅ 修复配置文件的解析异常

### 📈 性能改进
- **响应速度**: 提升系统整体响应速度
- **资源利用**: 优化CPU和内存资源利用率
- **并发性能**: 改进高并发场景下的性能表现
- **网络效率**: 优化网络连接和数据传输效率

### 🔧 使用改进
```bash
# 使用新的配置展示功能
./v2ray-manager show-config

# 自定义测试参数
./v2ray-manager speed-test-custom https://your-subscription-url \
  --timeout=60 \
  --test-url=https://www.google.com \
  --show-config

# Windows用户可以享受更好的并发控制
./v2ray-manager.exe speed-test https://your-subscription-url --concurrency=50
```

### 💔 破坏性变更
- 无破坏性变更，完全向后兼容

### 🎯 下个版本计划
- GUI界面开发
- 更多协议支持（VMess、Trojan）
- 配置文件管理系统
- 性能监控和分析工具

---

## v2.0.0 - 重大版本升级：增强清理系统、MVP测试器和双进程代理架构 (2025-01-20)

### 🚀 重大架构升级

#### 增强的清理系统
- **全新进程同步机制**: 实现更严格的进程停止和清理验证流程
- **智能清理逻辑**: 优化清理工具，增强进程同步机制和清理逻辑
- **重试机制**: 临时文件清理支持重试机制，确保清理结果的完整性和可靠性
- **深度清理**: 在AutoProxyManager和MVPTester中实现完善的资源回收

#### MVP测试器系统
- **新增MVP测试功能**: 添加MVP测试器支持，提供更专业的测试能力
- **测试工作流优化**: 改进测试流程，提高测试准确性和效率
- **状态管理增强**: 优化测试状态的跟踪和管理

#### 双进程代理系统
- **双进程架构**: 支持更复杂的代理管理场景
- **进程间通信**: 优化进程间的协调和通信机制
- **资源隔离**: 提高不同代理进程间的资源隔离度

### 🧹 系统清理功能重大升级

#### 临时文件清理增强
- **支持test_proxy_*文件清理**: 在多个工作流中添加对test_proxy_*文件的清理支持
- **状态文件管理**: 优化临时文件和状态文件的清理逻辑
- **清理效率提升**: 提升资源管理效率，减少残留文件

#### 进程管理优化
- **严格的进程验证**: 实现更严格的进程停止验证流程
- **同步机制**: 增强进程同步机制，避免竞态条件
- **资源回收**: 完善的资源回收机制，确保系统稳定性

### 📚 文档和用户体验

#### 文档更新
- **README增强**: 更新README文档，增加系统清理工具的使用说明
- **清理指南**: 提供详细的清理工具使用指南
- **架构说明**: 添加新架构的使用说明和最佳实践

### 🔧 技术改进

#### 依赖管理
- **fsnotify支持**: 新增fsnotify依赖，支持文件系统监控
- **系统依赖优化**: 更新sys依赖，提供更好的系统集成

#### 构建优化
- **编译优化**: 使用优化的编译参数，减小二进制文件大小
- **跨平台支持**: 完整支持Windows、Linux、macOS的多种架构
- **发布流程**: 完善的自动化发布流程

### 🐛 修复的问题

#### 高优先级修复
- ✅ 修复清理工具的进程同步问题
- ✅ 修复临时文件清理不完全的问题
- ✅ 修复MVP测试器的资源管理问题
- ✅ 修复双进程代理的状态同步问题

#### 稳定性改进
- ✅ 增强系统清理的可靠性
- ✅ 改进资源管理和回收机制
- ✅ 优化进程生命周期管理
- ✅ 提高系统整体稳定性

### 📈 性能改进
- **清理效率**: 显著提升清理工具的执行效率
- **资源管理**: 更高效的资源管理和回收机制
- **系统响应**: 提高系统整体响应速度和稳定性

### 💔 破坏性变更
- **架构升级**: 内部架构有重大改进，但保持API兼容性
- **清理行为**: 清理工具的行为更加严格和完善
- **配置变更**: 部分内部配置结构有调整

### 🎯 升级指南
```bash
# 1. 下载新版本
wget https://github.com/yxhpy/v2ray-subscription-manager/releases/download/v2.0.0/v2ray-subscription-manager-v2.0.0-linux-amd64.zip

# 2. 解压并替换
unzip v2ray-subscription-manager-v2.0.0-linux-amd64.zip
chmod +x v2ray-subscription-manager-linux-amd64

# 3. 清理旧版本的临时文件（推荐）
./v2ray-subscription-manager cleanup

# 4. 验证新功能
./v2ray-subscription-manager --version
```

### 🔮 下个版本计划
- GUI界面开发
- 更多协议支持（VMess、Trojan）
- 性能监控和分析功能
- 配置文件管理增强

---

## v1.2.0 - Windows兼容性重大修复和智能解码系统 (2025-01-XX)

### 🔧 Windows兼容性重大修复

#### Hysteria2并发下载修复
- **问题**: 在高并发测试环境下，多个worker同时检测到Hysteria2未安装，导致重复下载和资源冲突
- **解决方案**: 添加全局互斥锁`hysteria2DownloadMutex`，实现`SafeDownloadHysteria2()`方法
- **影响**: 完全解决并发下载问题，确保只有一个进程执行下载操作

#### Windows路径兼容性修复
- **问题**: Windows下V2Ray和Hysteria2可执行文件缺少`.exe`扩展名，导致"executable file not found in %PATH%"错误
- **解决方案**: 
  - 修改`NewHysteria2Downloader()`自动添加`.exe`扩展名
  - 改进`CheckHysteria2Installed()`支持检测和重命名已下载文件
  - 修复`proxy.go`中V2Ray启动路径，Windows下使用`v2ray.exe`
- **影响**: 完全解决Windows下的可执行文件路径问题

#### PE文件验证系统
- **问题**: 下载的Windows可执行文件报错`%1 is not a valid Win32 application`
- **解决方案**: 添加`validateWindowsExecutable()`方法验证PE文件格式
- **影响**: 确保下载的Windows可执行文件格式正确

### 🧠 智能解码系统

#### Base64智能检测
- **问题**: 订阅链接内容不是Base64编码时解码失败
- **解决方案**: 重写`decodeBase64()`函数，智能检测内容是否为Base64编码
- **特性**:
  - 自动检测协议前缀（如`://`）
  - 验证Base64字符集合法性
  - 智能回退到原始内容
- **影响**: 大幅提高订阅链接解析成功率

### 🧹 完善的临时文件清理系统

#### 跨平台清理策略
- **问题**: Hysteria2临时配置文件在测试结束后未完全清理，特别是Windows环境
- **解决方案**:
  - 为每个Hysteria2实例生成唯一配置文件路径（纳秒时间戳）
  - 实现`cleanupHysteria2TempFiles()`方法
  - 添加`cleanupWindowsHysteria2Files()`处理Windows文件锁
  - 支持多种文件名模式匹配

#### Windows文件锁处理
- **特性**: 
  - 多次重试删除机制（最多3次）
  - 文件句柄释放等待
  - 智能临时文件识别
- **影响**: 解决Windows下文件锁定导致的清理失败问题

### 🔄 下载重试机制

#### 文件完整性验证
- **新增**: 文件大小验证和完整性检查
- **新增**: 下载失败时自动重试（最多3次）
- **准备**: 多下载源支持框架
- **影响**: 提高下载成功率和文件可靠性

### 🐛 修复的问题

#### 高优先级修复
- ✅ 修复Hysteria2重复下载导致的资源冲突
- ✅ 修复Windows下可执行文件路径问题
- ✅ 修复Base64解码失败导致的订阅解析错误
- ✅ 修复临时配置文件清理不完全的问题
- ✅ 修复Windows PE文件验证问题

#### 稳定性改进
- ✅ 增强并发安全性，防止竞态条件
- ✅ 改进错误处理和日志输出
- ✅ 优化资源管理和清理流程
- ✅ 增强跨平台兼容性

### 📈 性能改进
- **并发安全**: 消除竞态条件，提高并发测试稳定性
- **资源管理**: 更完善的临时文件清理，减少磁盘占用
- **错误恢复**: 增强的重试机制，提高操作成功率

### 🔧 使用改进
```bash
# 现在可以在Windows下正常使用所有功能
v2ray-subscription-manager.exe speed-test https://your-subscription-url

# 支持非Base64编码的订阅链接
v2ray-subscription-manager parse https://plain-text-subscription-url

# 更好的并发测试体验
v2ray-subscription-manager speed-test-custom https://your-subscription-url --concurrency=100
```

### 💔 破坏性变更
- 无破坏性变更，完全向后兼容

### 🎯 下个版本计划
- VMess协议支持
- Trojan协议支持
- GUI界面开发
- 更多下载源支持

---

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