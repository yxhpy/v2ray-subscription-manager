# 临时文件清理指南

## 问题描述

在使用auto-proxy功能时，系统会创建一些临时文件和状态文件：
- `auto_proxy_best_node.json` - 存储最佳节点信息
- `test_proxy_*.json` - 节点测试时的临时配置文件
- `auto_proxy_state.json` - 自动代理状态文件
- `valid_nodes.json` - 有效节点缓存文件
- 其他临时配置文件

之前版本中，当auto-proxy停止时，这些文件没有被完全清理，可能会遗留在工作目录中。

## 修复内容

### 1. 增强的清理逻辑

在以下模块中增加了完善的文件清理机制：

- **AutoProxyManager** (`internal/core/workflow/auto_proxy.go`)
  - 停止时自动清理所有相关文件
  - 清理最佳节点文件、状态文件、有效节点文件

- **ProxyServer** (`internal/core/workflow/proxy_server.go`)
  - 清理测试代理时创建的临时文件
  - 包括 `test_proxy_*.json` 和 `test_proxy_*.yaml` 文件

- **MVPTester** (`internal/core/workflow/mvp_tester.go`)
  - 停止时清理状态文件和临时配置文件

- **SpeedTestWorkflow** (`internal/core/workflow/speedtest.go`)
  - 增强临时文件清理模式

### 2. 通用清理工具

创建了通用的清理工具模块 (`internal/utils/cleanup.go`)，包含：

- `CleanupTempFiles()` - 清理所有临时文件
- `CleanupHysteria2TempFiles()` - 清理Hysteria2相关临时文件
- `CleanupAutoProxyFiles()` - 清理Auto-proxy相关文件
- `ForceCleanupAll()` - 强制清理所有文件

### 3. 独立清理命令

提供了独立的清理工具 (`cmd/cleanup.go`)，用户可以手动运行：

```bash
# 构建清理工具
./scripts/build_cleanup.sh

# 运行清理
./bin/cleanup
```

## 清理的文件类型

### V2Ray相关临时文件
- `temp_v2ray_config_*.json`
- `test_v2ray_config_*.json`
- `proxy_server_v2ray_*.json`
- `temp_config_*.json`
- `test_proxy_*.json`

### Hysteria2相关临时文件
- `./hysteria2/test_config_*.yaml`
- `./hysteria2/temp_*.yaml`
- `./hysteria2/config_*.yaml`
- `./hysteria2/test_proxy_*.yaml`
- `./hysteria2/proxy_server_*.yaml`

### Auto-proxy相关文件
- `auto_proxy_best_node.json`
- `auto_proxy_state.json`
- `valid_nodes.json`
- `mvp_best_node.json`
- `proxy_state.json`

### 通用临时文件
- `*.tmp`
- `*.temp`

## 使用方法

### 自动清理
当auto-proxy正常停止时（通过Ctrl+C或程序结束），会自动执行清理。

### 手动清理
如果需要手动清理残留文件：

```bash
# 方法1：使用独立清理工具
./bin/cleanup

# 方法2：如果没有构建清理工具，可以手动删除
rm -f auto_proxy_best_node.json
rm -f test_proxy_*.json
rm -f auto_proxy_state.json
rm -f valid_nodes.json
rm -f mvp_best_node.json
rm -f proxy_state.json
rm -f temp_*.json
rm -f *.tmp *.temp
rm -f ./hysteria2/test_*.yaml
rm -f ./hysteria2/config_*.yaml
```

## 注意事项

1. **重要配置文件保护**: 清理过程会跳过重要的配置文件（如 `./hysteria2/config.yaml` 主配置文件）

2. **重启后恢复**: 清理后，重新启动auto-proxy服务会自动重新创建必要的文件

3. **进程清理**: 除了文件清理，停止时还会清理相关的v2ray和hysteria2进程

4. **安全性**: 清理操作只删除明确标识的临时文件，不会误删用户数据

## 故障排除

如果遇到文件清理问题：

1. **权限问题**: 确保对工作目录有写权限
2. **文件占用**: 确保相关进程已完全停止
3. **手动清理**: 使用 `./bin/cleanup` 工具强制清理
4. **重新构建**: 如果清理工具有问题，重新运行 `./scripts/build_cleanup.sh`

## 版本历史

- **v1.1**: 增加完善的文件清理机制
- **v1.0**: 初始版本，存在文件清理不完整的问题 