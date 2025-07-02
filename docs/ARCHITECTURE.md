# 项目架构文档

## 概述

本项目采用现代化的分层架构设计，遵循SOLID原则，使用依赖注入和接口分离等设计模式，确保代码的可维护性、可扩展性和可测试性。

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        Web UI Layer                         │
│                     (HTTP Handlers)                         │
├─────────────────────────────────────────────────────────────┤
│                      Service Layer                          │
│              (Business Logic & Interfaces)                  │
├─────────────────────────────────────────────────────────────┤
│                       Core Layer                            │
│                (Proxy Management & Parsing)                 │
├─────────────────────────────────────────────────────────────┤
│                      Platform Layer                         │
│                  (OS-specific Operations)                   │
└─────────────────────────────────────────────────────────────┘
```

## 目录结构

```
cmd/web-ui/
├── main.go                 # 应用程序入口点，依赖注入容器
├── handlers/               # HTTP处理器层
│   ├── subscription_handler.go  # 订阅管理处理器
│   ├── node_handler.go          # 节点管理处理器
│   ├── proxy_handler.go         # 代理管理处理器
│   └── status_handler.go        # 状态查询处理器
├── services/               # 业务逻辑服务层
│   ├── interfaces.go           # 服务接口定义
│   ├── subscription_service.go # 订阅服务实现
│   ├── node_service.go         # 节点服务实现
│   ├── proxy_service.go        # 代理服务实现
│   ├── system_service.go       # 系统服务实现
│   └── template_service.go     # 模板服务实现
├── models/                 # 数据模型层
│   └── models.go              # 统一数据模型定义
├── middleware/             # HTTP中间件
│   └── cors.go               # CORS处理中间件
└── templates/              # HTML模板
    └── index.html            # 主页模板
```

## 设计模式

### 1. 依赖注入模式 (Dependency Injection)

- **位置**: `main.go` 中的 `NewWebUIServer()`
- **作用**: 解耦组件之间的依赖关系，便于测试和维护
- **实现**: 通过构造函数注入依赖

```go
// 服务层依赖注入
s.subscriptionService = services.NewSubscriptionService()
s.proxyService = services.NewProxyService()
s.nodeService = services.NewNodeService(s.subscriptionService, s.proxyService)

// 处理器层依赖注入
s.subscriptionHandler = handlers.NewSubscriptionHandler(s.subscriptionService)
```

### 2. 接口分离原则 (Interface Segregation)

- **位置**: `services/interfaces.go`
- **作用**: 定义细粒度的接口，避免接口污染
- **实现**: 每个服务都有专门的接口定义

```go
type SubscriptionService interface {
    AddSubscription(url, name string) (*models.Subscription, error)
    GetAllSubscriptions() []*models.Subscription
    // ...
}
```

### 3. 策略模式 (Strategy Pattern)

- **位置**: `services/proxy_service.go`
- **作用**: 根据不同的代理协议使用不同的处理策略
- **实现**: 根据节点协议选择V2Ray或Hysteria2代理

```go
if node.Protocol == "hysteria2" {
    return n.proxyService.StartHysteria2Proxy(node)
} else {
    return n.proxyService.StartV2RayProxy(node)
}
```

### 4. 工厂模式 (Factory Pattern)

- **位置**: 各个 `New*Service()` 函数
- **作用**: 封装对象创建逻辑，统一管理依赖
- **实现**: 提供统一的服务创建接口

### 5. MVC模式 (Model-View-Controller)

- **Model**: `models/` 目录下的数据模型
- **View**: `templates/` 目录下的HTML模板
- **Controller**: `handlers/` 目录下的HTTP处理器

## 核心组件

### 1. 服务层 (Service Layer)

#### SubscriptionService
- **职责**: 订阅的增删改查、解析、测试
- **特点**: 线程安全、支持并发操作
- **依赖**: parser包用于订阅解析

#### NodeService  
- **职责**: 节点连接、测试、速度测试
- **特点**: 支持多种连接模式（随机端口、固定端口）
- **依赖**: SubscriptionService、ProxyService

#### ProxyService
- **职责**: 代理的启动、停止、状态管理
- **特点**: 支持V2Ray和Hysteria2两种代理类型
- **依赖**: internal/core/proxy包

#### SystemService
- **职责**: 系统状态查询、设置管理
- **特点**: 提供系统级别的配置和状态信息

### 2. 处理器层 (Handler Layer)

#### 统一的错误处理
- 所有处理器都使用统一的 `APIResponse` 格式
- 标准化的错误信息返回
- 一致的JSON响应格式

#### 请求验证
- 参数校验和类型检查
- 业务逻辑验证
- 错误信息本地化

### 3. 模型层 (Model Layer)

#### 统一的数据结构
- `APIResponse`: 标准API响应格式
- `Subscription`: 订阅信息模型
- `NodeTestResult`: 节点测试结果
- `ProxyStatus`: 代理状态信息

## 扩展性设计

### 1. 新增代理协议支持

1. 在 `internal/core/proxy/` 中添加新的代理管理器
2. 在 `ProxyService` 中添加对应的方法
3. 在 `NodeService` 中添加协议判断逻辑

### 2. 新增API功能

1. 在对应的 `Service` 接口中添加方法定义
2. 在服务实现中添加具体逻辑
3. 在 `Handler` 中添加HTTP处理方法
4. 在 `main.go` 中注册路由

### 3. 新增中间件

1. 在 `middleware/` 目录中创建新的中间件
2. 在路由注册时应用中间件

## 测试策略

### 1. 单元测试
- 每个服务都可以独立测试
- 使用接口mock依赖
- 测试覆盖率目标：80%+

### 2. 集成测试
- API端到端测试
- 数据库集成测试
- 外部服务集成测试

### 3. 性能测试
- 并发订阅解析测试
- 代理连接性能测试
- 内存泄漏检测

## 部署和运维

### 1. 构建
```bash
cd cmd/web-ui
go build -o ../../bin/webui .
```

### 2. 运行
```bash
./bin/webui
```

### 3. 配置
- 默认端口：8888
- 静态文件目录：web/static/
- 模板目录：cmd/web-ui/templates/

## 性能优化

### 1. 内存管理
- 使用对象池减少GC压力
- 及时释放不需要的资源
- 控制并发连接数

### 2. 并发优化
- 使用sync.RWMutex保护共享资源
- 合理使用goroutine
- 避免死锁和竞态条件

### 3. 网络优化
- 连接复用
- 超时控制
- 错误重试机制

## 安全考虑

### 1. 输入验证
- URL格式验证
- 端口范围检查
- 参数长度限制

### 2. 资源保护
- 防止资源耗尽攻击
- 限制并发请求数
- 内存使用监控

### 3. 错误处理
- 不暴露敏感信息
- 统一错误格式
- 日志记录和监控

## 未来规划

### 1. 短期目标
- 添加更多代理协议支持
- 完善测试覆盖率
- 优化用户界面

### 2. 长期目标
- 支持集群部署
- 添加用户认证
- 实现配置持久化
- 支持插件机制 