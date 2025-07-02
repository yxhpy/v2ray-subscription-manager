# 节点连接独立性问题修复

## 问题描述

在 Web UI 中，当用户点击多个节点进行连接时，每个新的连接会导致之前的连接断开失败。这是因为所有连接共享同一个代理管理器实例，每次新连接都会停止现有的代理。

## 问题根源

**原始架构问题：**

1. **全局管理器共享**：`NodeServiceImpl` 中有两个全局的代理管理器
   ```go
   type NodeServiceImpl struct {
       v2rayManager     *proxy.ProxyManager
       hysteria2Manager *proxy.Hysteria2ProxyManager
       // ...
   }
   ```

2. **相互干扰**：每次启动新连接时都会停止现有代理
   ```go
   func (n *NodeServiceImpl) startProxyForNode(node *types.Node, httpPort, socksPort int) (int, int, error) {
       // 停止现有代理 - 这导致了问题！
       n.v2rayManager.StopProxy()
       n.hysteria2Manager.StopHysteria2Proxy()
       // ...
   }
   ```

3. **实例重写**：虽然创建了新的管理器实例，但最后又赋值给全局变量
   ```go
   // 更新实例引用 - 这会影响其他连接！
   n.v2rayManager = manager
   ```

## 解决方案

### 1. 独立连接管理

为每个节点连接维护独立的代理管理器：

```go
type NodeServiceImpl struct {
    // 节点连接管理 - 每个连接独立的代理管理器
    nodeConnections map[string]*NodeConnection // key: subscriptionID_nodeIndex
    connectionMutex sync.RWMutex
    // ...
}

type NodeConnection struct {
    V2RayManager     *proxy.ProxyManager
    Hysteria2Manager *proxy.Hysteria2ProxyManager
    HTTPPort         int
    SOCKSPort        int
    Protocol         string
    IsActive         bool
}
```

### 2. 移除全局停止逻辑

新的 `startProxyForNodeWithConnection` 方法不会停止其他连接：

```go
func (n *NodeServiceImpl) startProxyForNodeWithConnection(subscriptionID string, nodeIndex int, node *types.Node, httpPort, socksPort int) (int, int, error) {
    // 只停止同一节点的现有连接
    n.removeNodeConnection(subscriptionID, nodeIndex)
    
    // 为每个连接创建独立的管理器实例
    v2rayManager := proxy.NewProxyManager()
    hysteria2Manager := proxy.NewHysteria2ProxyManager()
    
    // ... 启动代理逻辑
    
    // 创建独立的连接记录
    connection := &NodeConnection{
        V2RayManager:     v2rayManager,
        Hysteria2Manager: hysteria2Manager,
        HTTPPort:         actualHTTPPort,
        SOCKSPort:        actualSOCKSPort,
        Protocol:         node.Protocol,
        IsActive:         true,
    }
    
    // 添加到连接管理
    n.addNodeConnection(subscriptionID, nodeIndex, connection)
    
    return actualHTTPPort, actualSOCKSPort, nil
}
```

### 3. 智能端口冲突处理

对于固定端口连接，只停止占用该端口的连接：

```go
// 检查端口是否被占用，如果被占用则停止占用该端口的连接
if n.isPortOccupied(fixedHTTPPort) {
    n.stopConnectionsByPort(fixedHTTPPort)  // 只停止冲突的连接
}
```

## 修复效果

**修复前：**
- 连接节点A → 成功
- 连接节点B → 节点A断开，节点B成功
- 连接节点C → 节点B断开，节点C成功

**修复后：**
- 连接节点A → 成功，获得端口8001
- 连接节点B → 成功，获得端口8002，节点A仍运行
- 连接节点C → 成功，获得端口8003，节点A和B仍运行

## 核心改进

1. **真正独立**：每个节点连接使用独立的代理管理器实例
2. **无干扰启动**：新连接不会影响现有连接
3. **智能端口管理**：只在端口冲突时才停止相关连接
4. **完整生命周期**：每个连接都有独立的启动、运行、停止状态

## 向后兼容性

- 保留了原有的API接口
- 保持了相同的响应格式
- 维护了节点状态管理机制
- 支持所有原有的连接类型（http_random, socks_random, http_fixed, socks_fixed, disable）

这个修复确保了每个节点连接都是真正独立的，用户可以同时连接多个节点而不会相互影响。 