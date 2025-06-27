# V2Ray订阅管理器 - 项目结构重构方案

## 当前项目结构问题

当前项目所有Go源代码文件都放在根目录下，缺乏明确的模块划分，不利于代码维护和扩展。

## 建议的新项目结构

```
verify_v2ray_ui2/
├── cmd/                          # 命令行入口
│   └── v2ray-manager/
│       └── main.go              # 主入口文件
│
├── internal/                     # 内部包（不对外暴露）
│   ├── core/                    # 核心业务逻辑
│   │   ├── downloader/          # 下载器模块
│   │   │   ├── v2ray.go         # V2Ray下载器
│   │   │   ├── hysteria2.go     # Hysteria2下载器
│   │   │   └── common.go        # 通用下载逻辑
│   │   │
│   │   ├── proxy/               # 代理管理模块
│   │   │   ├── v2ray.go         # V2Ray代理管理
│   │   │   ├── hysteria2.go     # Hysteria2代理管理
│   │   │   ├── manager.go       # 代理管理器
│   │   │   └── status.go        # 状态管理
│   │   │
│   │   ├── parser/              # 订阅解析模块
│   │   │   ├── subscription.go  # 订阅链接解析
│   │   │   ├── node.go          # 节点数据结构
│   │   │   └── protocols/       # 协议解析
│   │   │       ├── vless.go     # VLESS协议
│   │   │       ├── shadowsocks.go # SS协议
│   │   │       └── hysteria2.go # Hysteria2协议
│   │   │
│   │   └── workflow/            # 工作流模块
│   │       ├── speedtest.go     # 测速工作流
│   │       └── config.go        # 工作流配置
│   │
│   ├── platform/                # 平台相关代码
│   │   ├── unix.go              # Unix系统特定代码
│   │   └── windows.go           # Windows系统特定代码
│   │
│   └── utils/                   # 工具函数
│       ├── network.go           # 网络工具
│       ├── file.go              # 文件操作
│       └── system.go            # 系统工具
│
├── pkg/                         # 可对外暴露的包
│   └── types/                   # 公共类型定义
│       ├── node.go              # 节点类型
│       ├── config.go            # 配置类型
│       └── status.go            # 状态类型
│
├── configs/                     # 配置文件
│   ├── v2ray/                   # V2Ray配置模板
│   └── hysteria2/               # Hysteria2配置模板
│
├── scripts/                     # 脚本文件
│   ├── build.sh                 # 构建脚本
│   ├── release.sh               # 发布脚本
│   └── release.bat              # Windows发布脚本
│
├── bin/                         # 编译产物和发布文件
│   └── releases/                # 发布版本
│
├── docs/                        # 文档
│   ├── README.md                # 主文档
│   ├── API.md                   # API文档
│   └── DEVELOPMENT.md           # 开发文档
│
├── test/                        # 测试文件
│   ├── integration/             # 集成测试
│   └── testdata/                # 测试数据
│
├── go.mod                       # Go模块文件
├── go.sum                       # Go依赖锁定文件
├── .gitignore                   # Git忽略文件
├── LICENSE                      # 许可证
└── Makefile                     # 构建配置
```

## 模块职责划分

### 1. cmd/ - 命令行入口
- 负责命令行参数解析
- 调用内部模块完成具体功能
- 保持简洁，主要是路由逻辑

### 2. internal/core/ - 核心业务逻辑
- **downloader/**: 负责V2Ray和Hysteria2的自动下载
- **proxy/**: 代理启动、停止、状态管理
- **parser/**: 订阅链接解析，支持多种协议
- **workflow/**: 测速等复杂工作流

### 3. internal/platform/ - 平台相关
- 分离Unix和Windows特定的系统调用
- 处理跨平台兼容性问题

### 4. internal/utils/ - 工具函数
- 网络请求、文件操作等通用工具
- 避免代码重复

### 5. pkg/ - 公共类型
- 定义可被外部引用的数据结构
- 保持API稳定性

### 6. configs/ - 配置管理
- V2Ray和Hysteria2的配置模板
- 支持配置自定义

## 重构的优势

1. **模块化清晰**: 每个模块职责明确，便于维护
2. **代码复用**: 公共逻辑提取到utils包
3. **测试友好**: 便于编写单元测试和集成测试
4. **扩展性好**: 新增协议或功能时结构清晰
5. **跨平台支持**: 平台相关代码分离
6. **文档完善**: 专门的docs目录管理文档

## 实施建议

1. **分阶段重构**: 先创建新目录结构，逐步迁移代码
2. **保持兼容**: 重构过程中保证功能不受影响
3. **添加测试**: 重构时补充单元测试
4. **更新文档**: 同步更新相关文档

这种结构更符合Go项目的最佳实践，便于后续维护和功能扩展。 