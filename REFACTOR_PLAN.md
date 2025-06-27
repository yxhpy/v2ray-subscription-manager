# V2Ray订阅管理器重构实施计划

## 项目现状分析

### 当前文件分布
- **Go源文件**: 9个文件，共3646行代码
- **主要模块**: 
  - `main.go` - 命令行入口和路由 (381行)
  - `workflow.go` - 测速工作流 (856行)
  - `proxy.go` - 代理管理 (772行)
  - `parser.go` - 订阅解析 (544行)
  - `downloader.go` - V2Ray下载器 (447行)
  - `hysteria2_downloader.go` - Hysteria2下载器 (432行)
  - `hysteria2_proxy.go` - Hysteria2代理管理 (196行)
  - `proxy_unix.go` - Unix平台特定代码 (14行)
  - `proxy_windows.go` - Windows平台特定代码 (13行)

### 问题分析
1. **单一职责原则违反**: 多个功能混合在同一文件中
2. **代码复用性差**: 相似逻辑在不同文件中重复
3. **测试困难**: 缺乏清晰的接口和模块边界
4. **扩展性差**: 新增功能需要修改多个文件

## 重构实施步骤

### 第一阶段：创建目录结构和迁移文件

```bash
# 1. 创建新的目录结构
make create-structure

# 2. 迁移脚本和文档文件
make migrate-files
```

### 第二阶段：重构Go代码文件

#### 2.1 公共类型定义 (pkg/types/)

**文件**: `pkg/types/node.go`
```go
// 节点相关类型定义
type Node struct {
    Name     string
    Protocol string
    Address  string
    Port     int
    // ... 其他字段
}

type NodeList []*Node
```

**文件**: `pkg/types/config.go`
```go
// 配置相关类型
type ProxyConfig struct {
    HTTPPort  int
    SOCKSPort int
    // ... 其他配置
}
```

**文件**: `pkg/types/status.go`
```go
// 状态相关类型
type ProxyStatus struct {
    Running   bool
    NodeName  string
    Protocol  string
    HTTPPort  int
    SOCKSPort int
}
```

#### 2.2 工具函数 (internal/utils/)

**文件**: `internal/utils/network.go`
```go
// 网络相关工具函数
func DownloadFile(url, dest string) error
func TestConnection(proxyURL string) error
func GetFreePort() (int, error)
```

**文件**: `internal/utils/file.go`
```go
// 文件操作工具函数
func EnsureDir(path string) error
func ExtractArchive(src, dest string) error
func SetExecutable(path string) error
```

**文件**: `internal/utils/system.go`
```go
// 系统相关工具函数
func GetOS() string
func GetArch() string
func KillProcess(pid int) error
```

#### 2.3 平台相关代码 (internal/platform/)

**文件**: `internal/platform/unix.go`
```go
//go:build !windows
// Unix系统特定实现
```

**文件**: `internal/platform/windows.go`
```go
//go:build windows
// Windows系统特定实现
```

#### 2.4 核心业务逻辑 (internal/core/)

**下载器模块** (`internal/core/downloader/`)
- `common.go` - 通用下载逻辑
- `v2ray.go` - V2Ray下载器
- `hysteria2.go` - Hysteria2下载器

**代理管理模块** (`internal/core/proxy/`)
- `manager.go` - 代理管理器接口
- `v2ray.go` - V2Ray代理实现
- `hysteria2.go` - Hysteria2代理实现
- `status.go` - 状态管理

**解析模块** (`internal/core/parser/`)
- `subscription.go` - 订阅链接解析
- `node.go` - 节点数据处理
- `protocols/vless.go` - VLESS协议解析
- `protocols/shadowsocks.go` - SS协议解析
- `protocols/hysteria2.go` - Hysteria2协议解析

**工作流模块** (`internal/core/workflow/`)
- `speedtest.go` - 测速工作流
- `config.go` - 工作流配置

#### 2.5 命令行入口 (cmd/v2ray-manager/)

**文件**: `cmd/v2ray-manager/main.go`
```go
// 简化的主入口，只负责命令路由
func main() {
    // 解析命令行参数
    // 调用对应的核心模块
}
```

### 第三阶段：重构具体实现

#### 3.1 代码迁移优先级

1. **高优先级** (核心功能，独立性强):
   - 公共类型定义 → `pkg/types/`
   - 工具函数 → `internal/utils/`
   - 平台相关代码 → `internal/platform/`

2. **中优先级** (业务逻辑，相互依赖):
   - 下载器模块 → `internal/core/downloader/`
   - 解析模块 → `internal/core/parser/`

3. **低优先级** (复杂模块，依赖较多):
   - 代理管理模块 → `internal/core/proxy/`
   - 工作流模块 → `internal/core/workflow/`
   - 命令行入口 → `cmd/v2ray-manager/`

#### 3.2 接口设计

```go
// 代理管理器接口
type ProxyManager interface {
    Start(node *Node) error
    Stop() error
    GetStatus() *ProxyStatus
    Test() error
}

// 下载器接口
type Downloader interface {
    Download() error
    Check() bool
    GetVersion() string
}

// 解析器接口
type Parser interface {
    Parse(url string) ([]*Node, error)
    ParseProtocol(uri string) (*Node, error)
}
```

### 第四阶段：测试和验证

#### 4.1 单元测试
```bash
# 为每个模块添加单元测试
mkdir -p internal/core/downloader/tests
mkdir -p internal/core/proxy/tests
mkdir -p internal/core/parser/tests
```

#### 4.2 集成测试
```bash
# 添加集成测试
mkdir -p test/integration
```

#### 4.3 功能验证
```bash
# 验证重构后功能完整性
make test
make build
./bin/v2ray-subscription-manager --help
```

## 重构执行命令

### 自动化命令序列
```bash
# 第一步：准备工作
make refactor-check
make create-structure
make migrate-files

# 第二步：代码重构 (需要手动执行)
# 按照上述优先级逐步迁移Go文件

# 第三步：测试验证
make fmt
make lint
make test
make build

# 第四步：最终检查
make status
```

### 风险控制

1. **备份当前代码**
```bash
git add .
git commit -m "重构前备份"
git tag backup-before-refactor
```

2. **分支开发**
```bash
git checkout -b refactor-project-structure
```

3. **渐进式重构**
- 每次只重构一个模块
- 重构后立即测试
- 确保功能正常后再继续

4. **回滚方案**
```bash
# 如果重构出现问题，可以回滚
git checkout main
git reset --hard backup-before-refactor
```

## 预期收益

1. **代码质量提升**: 模块化清晰，职责分明
2. **维护性增强**: 便于定位和修复问题
3. **扩展性改善**: 新增功能时结构清晰
4. **测试友好**: 便于编写和维护测试
5. **团队协作**: 多人开发时冲突减少

## 时间估算

- **第一阶段**: 0.5天 (目录创建和文件迁移)
- **第二阶段**: 2-3天 (Go代码重构)
- **第三阶段**: 1-2天 (测试和验证)
- **第四阶段**: 0.5天 (文档更新)

**总计**: 4-6天

这个重构计划将显著提升项目的代码质量和可维护性，为后续功能扩展打下良好基础。 