# V2Ray订阅管理器 Makefile

# 项目信息
PROJECT_NAME = v2ray-subscription-manager
BINARY_NAME = v2ray-subscription-manager
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH = $(shell git rev-parse --short HEAD)

# Go相关配置
GO = go
GOFLAGS = -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 目录配置
BIN_DIR = bin
SCRIPTS_DIR = scripts
DOCS_DIR = docs
CONFIG_DIR = configs
TEST_DIR = test

# 默认目标
.PHONY: help
help: ## 显示帮助信息
	@echo "V2Ray订阅管理器 - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 构建相关
.PHONY: build
build: ## 构建项目
	@echo "构建 $(PROJECT_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) .

.PHONY: build-all
build-all: ## 构建所有平台版本
	@echo "构建所有平台版本..."
	@chmod +x $(SCRIPTS_DIR)/build.sh
	@$(SCRIPTS_DIR)/build.sh

.PHONY: clean
clean: ## 清理构建产物
	@echo "清理构建产物..."
	@rm -rf $(BIN_DIR)/$(BINARY_NAME)*
	@rm -f *.log *.json *.txt

# 开发相关
.PHONY: dev
dev: build ## 开发模式构建
	@echo "开发构建完成: $(BIN_DIR)/$(BINARY_NAME)"

.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "生成测试覆盖率报告..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

.PHONY: lint
lint: ## 代码检查
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GO) fmt ./...

.PHONY: mod
mod: ## 整理依赖
	@echo "整理Go模块依赖..."
	$(GO) mod tidy
	$(GO) mod verify

# 项目重构相关
.PHONY: create-structure
create-structure: ## 创建新的项目目录结构
	@echo "创建新的项目目录结构..."
	@mkdir -p cmd/v2ray-manager
	@mkdir -p internal/core/downloader
	@mkdir -p internal/core/proxy
	@mkdir -p internal/core/parser/protocols
	@mkdir -p internal/core/workflow
	@mkdir -p internal/platform
	@mkdir -p internal/utils
	@mkdir -p pkg/types
	@mkdir -p $(CONFIG_DIR)/v2ray
	@mkdir -p $(CONFIG_DIR)/hysteria2
	@mkdir -p $(SCRIPTS_DIR)
	@mkdir -p $(DOCS_DIR)
	@mkdir -p $(TEST_DIR)/integration
	@mkdir -p $(TEST_DIR)/testdata
	@mkdir -p $(BIN_DIR)/releases
	@echo "目录结构创建完成"

.PHONY: migrate-files
migrate-files: create-structure ## 迁移现有文件到新结构
	@echo "迁移文件到新结构..."
	# 移动脚本文件
	@if [ -f build.sh ]; then mv build.sh $(SCRIPTS_DIR)/; fi
	@if [ -f release.sh ]; then mv release.sh $(SCRIPTS_DIR)/; fi
	@if [ -f release.bat ]; then mv release.bat $(SCRIPTS_DIR)/; fi
	# 移动文档
	@if [ -f README.md ]; then cp README.md $(DOCS_DIR)/; fi
	@if [ -f RELEASE_NOTES.md ]; then mv RELEASE_NOTES.md $(DOCS_DIR)/; fi
	@echo "文件迁移完成"

.PHONY: refactor-check
refactor-check: ## 检查重构准备情况
	@echo "检查重构准备情况..."
	@echo "当前Go文件:"
	@find . -name "*.go" -not -path "./internal/*" -not -path "./cmd/*" -not -path "./pkg/*" | head -10
	@echo ""
	@echo "建议按以下顺序重构:"
	@echo "1. make create-structure  # 创建目录结构"
	@echo "2. make migrate-files     # 迁移文件"
	@echo "3. 手动重构Go代码文件"
	@echo "4. make test             # 测试重构结果"

# 发布相关
.PHONY: release
release: ## 创建发布版本
	@echo "创建发布版本..."
	@if [ -z "$(VERSION)" ]; then \
		echo "请指定版本号: make release VERSION=v1.x.x"; \
		exit 1; \
	fi
	@chmod +x $(SCRIPTS_DIR)/release.sh
	@$(SCRIPTS_DIR)/release.sh $(VERSION)

# 工具相关
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发工具安装完成"

.PHONY: check-deps
check-deps: ## 检查依赖
	@echo "检查项目依赖..."
	$(GO) list -m -u all

.PHONY: update-deps
update-deps: ## 更新依赖
	@echo "更新项目依赖..."
	$(GO) get -u ./...
	$(GO) mod tidy

# 运行相关
.PHONY: run
run: build ## 构建并运行程序
	@echo "运行程序..."
	@./$(BIN_DIR)/$(BINARY_NAME)

.PHONY: run-help
run-help: build ## 显示程序帮助
	@./$(BIN_DIR)/$(BINARY_NAME) --help || ./$(BIN_DIR)/$(BINARY_NAME) help || ./$(BIN_DIR)/$(BINARY_NAME)

# 清理相关
.PHONY: clean-all
clean-all: clean ## 深度清理
	@echo "深度清理..."
	@rm -f coverage.out coverage.html
	@rm -f *.log *.json *.txt
	@$(GO) clean -cache
	@$(GO) clean -modcache

.PHONY: status
status: ## 显示项目状态
	@echo "=== 项目状态 ==="
	@echo "项目名称: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "提交哈希: $(COMMIT_HASH)"
	@echo "Go版本: $(shell $(GO) version)"
	@echo "目标平台: $(GOOS)/$(GOARCH)"
	@echo ""
	@echo "=== Git状态 ==="
	@git status --short || echo "不是Git仓库"
	@echo ""
	@echo "=== 文件统计 ==="
	@echo "Go文件数量: $(shell find . -name '*.go' | wc -l)"
	@echo "总代码行数: $(shell find . -name '*.go' -exec wc -l {} + | tail -1 | awk '{print $$1}')"

# 设置默认目标
.DEFAULT_GOAL := help 