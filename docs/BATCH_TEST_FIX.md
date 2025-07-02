# 批量测试连接问题修复

## 问题描述

在Web UI的批量测试功能中，经常出现以下错误：
```
连接测试: HTTP代理连接失败: dial tcp 127.0.0.1:端口 connect: connection refused
```

## 问题分析

经过分析，发现问题的根本原因是：

1. **端口分配冲突**：批量测试时多个节点可能尝试使用相同或相近的端口
2. **代理启动时间不足**：代理程序还没完全启动就开始连接测试
3. **并发控制不足**：过多并发导致系统资源竞争
4. **清理机制不完善**：临时代理管理器没有正确清理，导致端口占用
5. **配置文件冲突**：多个代理实例使用相同的配置文件路径，导致配置覆盖

## 修复方案

### 1. 优化端口分配策略

**修改前**：
```go
portBase := int(atomic.AddInt64(&n.portCounter, 10))
httpPort := portBase
socksPort := portBase + 1
```

**修改后**：
```go
// 获取唯一端口号，增加更大的间隔避免冲突
portBase := int(atomic.AddInt64(&n.portCounter, 20))
httpPort := portBase
socksPort := portBase + 1

// 确保端口可用
for i := 0; i < 10; i++ {
    if n.isPortAvailable(httpPort) && n.isPortAvailable(socksPort) {
        break
    }
    portBase = int(atomic.AddInt64(&n.portCounter, 20))
    httpPort = portBase
    socksPort = portBase + 1
}
```

### 2. 增加端口可用性检查

新增 `isPortAvailable` 方法：
```go
func (n *NodeServiceImpl) isPortAvailable(port int) bool {
    conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    conn.Close()
    return true
}
```

### 3. 改进代理启动等待机制

**修改前**：
```go
time.Sleep(3 * time.Second)
return tempManager.TestProxy()
```

**修改后**：
```go
// 等待代理启动，增加等待时间确保稳定
time.Sleep(5 * time.Second)

// 验证代理是否真正运行
if !tempManager.IsRunning() {
    return fmt.Errorf("代理启动后未能正常运行")
}

return tempManager.TestProxy()
```

### 4. 优化批量测试并发控制

**修改前**：
```go
maxConcurrency := 3 // 限制并发数量
```

**修改后**：
```go
// 减少并发数量以避免端口冲突和资源竞争
maxConcurrency := 2 // 进一步减少并发数量

defer func() {
    <-semaphore
    // 在每个任务完成后稍作延迟，给系统清理时间
    time.Sleep(500 * time.Millisecond)
}()
```

### 5. 改进资源清理机制

**修改前**：
```go
defer tempManager.StopProxy()
```

**修改后**：
```go
defer func() {
    // 确保清理代理
    tempManager.StopProxy()
    // 给清理一些时间
    time.Sleep(1 * time.Second)
}()
```

### 6. 确保配置文件唯一性

**修改前**：
```go
// V2Ray - 所有实例使用相同配置文件
ConfigPath: "temp_v2ray_config.json",

// Hysteria2 - 已有唯一配置，但可以更优化
downloader.ConfigPath = fmt.Sprintf("./hysteria2/config_%d.yaml", time.Now().UnixNano())
```

**修改后**：
```go
// V2Ray - 每个实例使用唯一配置文件
uniqueConfigPath := fmt.Sprintf("temp_v2ray_config_%d.json", time.Now().UnixNano())

// 专门的测试实例构造方法
func NewTestProxyManager() *ProxyManager {
    uniqueConfigPath := fmt.Sprintf("test_v2ray_config_%d_%d.json", time.Now().UnixNano(), os.Getpid())
    // ...
}

// Hysteria2 - 测试专用配置文件
func NewTestHysteria2ProxyManager() *Hysteria2ProxyManager {
    downloader.ConfigPath = fmt.Sprintf("./hysteria2/test_config_%d_%d.yaml", time.Now().UnixNano(), os.Getpid())
    // ...
}
```

### 7. 增加详细日志输出

```go
fmt.Printf("🚀 开始批量测试 %d 个节点，并发数: %d\n", len(nodeIndexes), maxConcurrency)
fmt.Printf("🧪 开始测试节点 %d\n", nodeIndex)
fmt.Printf("✅ 节点 %d 测试完成: %s\n", nodeIndex, result.NodeName)
fmt.Printf("❌ 节点 %d 测试失败: %s\n", nodeIndex, err.Error())
fmt.Printf("📊 批量测试完成，共测试 %d 个节点\n", len(results))
```

## 修复效果

### 修复前的问题
- 高概率出现端口连接失败
- 批量测试经常失败
- 错误信息：`dial tcp 127.0.0.1:端口 connect: connection refused`

### 修复后的改进
- 显著降低端口冲突概率
- 增加代理启动可靠性
- 更好的错误处理和日志输出
- 改进的资源清理机制
- 确保配置文件唯一性，避免配置覆盖

## 测试验证

使用提供的测试脚本进行验证：

```bash
./test_batch_fix.sh
```

测试步骤：
1. 启动Web UI服务器
2. 添加订阅链接
3. 选择多个节点进行批量测试
4. 观察是否还出现端口连接错误

## 技术细节

### 配置文件管理
- V2Ray配置：`test_v2ray_config_{时间戳}_{进程ID}.json`
- Hysteria2配置：`./hysteria2/test_config_{时间戳}_{进程ID}.yaml`
- 自动清理：停止代理时自动删除配置文件
- 清理脚本：`scripts/cleanup_temp_configs.sh`

### 端口分配范围
- 起始端口：9000
- 间隔：20（确保HTTP和SOCKS端口不冲突）
- 重试机制：最多10次端口查找

### 并发控制
- 最大并发数：2
- 任务间延迟：500ms
- 清理延迟：1秒

### 等待时间优化
- 代理启动等待：5秒（从3秒增加）
- 额外状态验证：确保代理真正运行

## 相关文件

修改的文件：
- `cmd/web-ui/services/node_service.go` - 主要修复逻辑
- `internal/core/proxy/v2ray.go` - V2Ray代理管理器配置文件唯一性
- `internal/core/proxy/hysteria2.go` - Hysteria2代理管理器配置文件唯一性
- `test_batch_fix.sh` - 测试验证脚本
- `scripts/cleanup_temp_configs.sh` - 临时文件清理脚本

新增的方法：
- `NewTestProxyManager()` - V2Ray测试专用代理管理器
- `NewTestHysteria2ProxyManager()` - Hysteria2测试专用代理管理器
- `isPortAvailable()` - 端口可用性检查

影响的方法：
- `BatchTestNodes` - 批量测试主逻辑
- `testV2RayNode` - V2Ray节点测试
- `testHysteria2Node` - Hysteria2节点测试
- `speedTestV2RayNode` - V2Ray速度测试
- `speedTestHysteria2Node` - Hysteria2速度测试 