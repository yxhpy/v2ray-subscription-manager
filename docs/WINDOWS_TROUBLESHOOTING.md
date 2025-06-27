# Windows环境故障排除指南

## 🔍 问题现象
Windows环境下所有节点测试失败，而macOS/Linux环境正常工作。

## 🛠️ 解决方案

### 1. 管理员权限
在Windows上运行程序需要管理员权限：
```cmd
# 以管理员身份运行CMD或PowerShell
.\bin\v2ray-subscription-manager.exe speed-test-custom "订阅链接"
```

### 2. 防火墙设置
Windows防火墙可能阻止V2Ray进程：
1. 打开"Windows Defender 防火墙"
2. 点击"允许应用或功能通过Windows Defender防火墙"
3. 添加V2Ray程序到例外列表

### 3. V2Ray版本检查
确保V2Ray版本兼容：
```cmd
# 检查V2Ray版本
.\v2ray\v2ray.exe version

# 推荐使用最新稳定版本
```

### 4. 网络代理设置
检查Windows系统代理设置：
1. 设置 → 网络和Internet → 代理
2. 确保"自动检测设置"已关闭
3. 确保"使用代理服务器"已关闭

### 5. 端口冲突检查
检查端口是否被占用：
```cmd
# 检查8080端口
netstat -ano | findstr :8080

# 检查1080端口  
netstat -ano | findstr :1080
```

### 6. 杀毒软件排除
将程序目录添加到杀毒软件白名单：
- V2Ray执行文件
- 配置文件目录
- 程序工作目录

### 7. 详细日志调试
启用详细日志模式：
```cmd
# 使用较低的并发数和较长的超时时间
.\bin\v2ray-subscription-manager.exe speed-test-custom "订阅链接" --concurrency=1 --timeout=30
```

### 8. 手动V2Ray测试
手动测试V2Ray配置：
```cmd
# 使用生成的配置文件手动启动V2Ray
.\v2ray\v2ray.exe run -config temp_config_xxxxx.json
```

## 🔧 推荐的Windows测试命令
```cmd
# 基础测试（单线程，长超时）
.\bin\v2ray-subscription-manager.exe speed-test-custom "订阅链接" --concurrency=1 --timeout=30 --test-url="http://www.baidu.com"

# 如果百度测试成功，再尝试Google
.\bin\v2ray-subscription-manager.exe speed-test-custom "订阅链接" --concurrency=1 --timeout=30 --test-url="https://www.google.com"
```

## 📝 已知限制
1. Windows环境下某些网络配置可能导致代理连接失败
2. 部分Windows版本的网络堆栈与V2Ray可能有兼容性问题
3. Windows防病毒软件可能干扰代理程序运行

## 💡 技术细节
- macOS环境测试成功率：77.2%（44/57节点）
- Windows环境测试成功率：0%（0/57节点）
- 问题不在于SS链接解析（已修复）
- 问题不在于加密方法支持（已完善）
- 问题在于Windows特定的网络环境配置 