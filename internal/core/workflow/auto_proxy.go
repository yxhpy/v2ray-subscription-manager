package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/core/downloader"
	"github.com/yxhpy/v2ray-subscription-manager/internal/utils"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// AutoProxyManager 双进程自动代理管理器
type AutoProxyManager struct {
	config types.AutoProxyConfig
	state  types.AutoProxyState

	// 测试进程相关
	tester       *MVPTester
	testerCtx    context.Context
	testerCancel context.CancelFunc

	// 代理服务进程相关
	proxyServer  *ProxyServer
	serverCtx    context.Context
	serverCancel context.CancelFunc

	// 通用管理
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
	bestNodeFile string

	// 用于进程间通信
	testResults    []types.ValidNode
	blacklist      map[string]time.Time
	blacklistMutex sync.RWMutex
}

// NewAutoProxyManager 创建新的双进程自动代理管理器
func NewAutoProxyManager(config types.AutoProxyConfig) *AutoProxyManager {
	ctx, cancel := context.WithCancel(context.Background())
	testerCtx, testerCancel := context.WithCancel(ctx)
	serverCtx, serverCancel := context.WithCancel(ctx)

	// 设置默认值
	if config.HTTPPort == 0 {
		config.HTTPPort = 7890
	}
	if config.SOCKSPort == 0 {
		config.SOCKSPort = 7891
	}
	if config.UpdateInterval == 0 {
		config.UpdateInterval = 10 * time.Minute
	}
	if config.TestConcurrency == 0 {
		config.TestConcurrency = 20
	}
	if config.TestTimeout == 0 {
		config.TestTimeout = 30 * time.Second
	}
	if config.TestURL == "" {
		config.TestURL = "http://www.google.com"
	}
	if config.MinPassingNodes == 0 {
		config.MinPassingNodes = 5
	}
	if config.StateFile == "" {
		config.StateFile = "./auto_proxy_state.json"
	}
	if config.ValidNodesFile == "" {
		config.ValidNodesFile = "./valid_nodes.json"
	}

	// 最佳节点文件路径
	bestNodeFile := "auto_proxy_best_node.json"

	// 创建MVP测试器
	tester := NewMVPTester(config.SubscriptionURL)
	tester.SetStateFile(bestNodeFile)
	tester.SetInterval(config.UpdateInterval)
	tester.SetMaxNodes(config.MaxNodes)
	tester.SetConcurrency(config.TestConcurrency)

	// 创建代理服务器
	proxyServer := NewProxyServer(bestNodeFile, config.HTTPPort, config.SOCKSPort)

	return &AutoProxyManager{
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		testerCtx:    testerCtx,
		testerCancel: testerCancel,
		serverCtx:    serverCtx,
		serverCancel: serverCancel,
		tester:       tester,
		proxyServer:  proxyServer,
		bestNodeFile: bestNodeFile,
		testResults:  make([]types.ValidNode, 0),
		state: types.AutoProxyState{
			Config:     config,
			ValidNodes: make([]types.ValidNode, 0),
			StartTime:  time.Now(),
		},
		blacklist: make(map[string]time.Time),
	}
}

// Start 启动双进程自动代理系统
func (m *AutoProxyManager) Start() error {
	fmt.Printf("🚀 启动双进程自动代理系统...\n")
	fmt.Printf("📡 订阅链接: %s\n", m.config.SubscriptionURL)
	fmt.Printf("🌐 HTTP代理: http://127.0.0.1:%d\n", m.config.HTTPPort)
	fmt.Printf("🧦 SOCKS代理: socks5://127.0.0.1:%d\n", m.config.SOCKSPort)
	fmt.Printf("⏰ 更新间隔: %v\n", m.config.UpdateInterval)
	fmt.Printf("📄 最佳节点文件: %s\n", m.bestNodeFile)

	// 设置信号处理
	m.setupSignalHandler()

	// 检查依赖
	if err := m.checkDependencies(); err != nil {
		return fmt.Errorf("依赖检查失败: %v", err)
	}

	// 启动状态
	m.state.Running = true
	m.state.StartTime = time.Now()

	// 启动进程1：节点测试器
	fmt.Printf("🧪 启动进程1：节点测试器...\n")
	go m.runTesterProcess()

	// 等待一下，让测试器先运行并生成初始的最佳节点文件
	time.Sleep(3 * time.Second)

	// 启动进程2：代理服务器
	fmt.Printf("🌐 启动进程2：代理服务器...\n")
	go m.runProxyServerProcess()

	// 启动监控协程
	go m.monitorProcesses()

	fmt.Printf("✅ 双进程自动代理系统启动成功！\n")
	fmt.Printf("📝 按 Ctrl+C 停止服务\n")

	return nil
}

// runTesterProcess 运行测试进程
func (m *AutoProxyManager) runTesterProcess() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ 测试进程异常退出: %v\n", r)
		}
	}()

	fmt.Printf("  🧪 测试进程启动中...\n")

	// 重写测试器的上下文
	m.tester.ctx = m.testerCtx
	m.tester.cancel = m.testerCancel

	if err := m.tester.Start(); err != nil {
		fmt.Printf("❌ 测试进程启动失败: %v\n", err)
	}
}

// runProxyServerProcess 运行代理服务进程
func (m *AutoProxyManager) runProxyServerProcess() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ 代理服务进程异常退出: %v\n", r)
		}
	}()

	fmt.Printf("  🌐 代理服务进程启动中...\n")

	// 重写代理服务器的上下文
	m.proxyServer.ctx = m.serverCtx
	m.proxyServer.cancel = m.serverCancel

	if err := m.proxyServer.Start(); err != nil {
		fmt.Printf("❌ 代理服务进程启动失败: %v\n", err)
	}
}

// monitorProcesses 监控进程状态
func (m *AutoProxyManager) monitorProcesses() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkProcessHealth()
		case <-m.ctx.Done():
			return
		}
	}
}

// checkProcessHealth 检查进程健康状态
func (m *AutoProxyManager) checkProcessHealth() {
	// 检查测试进程状态
	select {
	case <-m.testerCtx.Done():
		fmt.Printf("⚠️ 检测到测试进程已停止，尝试重启...\n")
		m.testerCtx, m.testerCancel = context.WithCancel(m.ctx)
		go m.runTesterProcess()
	default:
		// 测试进程正常运行
	}

	// 检查代理服务进程状态
	select {
	case <-m.serverCtx.Done():
		fmt.Printf("⚠️ 检测到代理服务进程已停止，尝试重启...\n")
		m.serverCtx, m.serverCancel = context.WithCancel(m.ctx)
		go m.runProxyServerProcess()
	default:
		// 代理服务进程正常运行
	}

	// 更新状态
	m.updateSystemStatus()
}

// updateSystemStatus 更新系统状态
func (m *AutoProxyManager) updateSystemStatus() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 从最佳节点文件读取当前状态
	if data, err := os.ReadFile(m.bestNodeFile); err == nil {
		var mvpState MVPState
		if err := json.Unmarshal(data, &mvpState); err == nil && mvpState.BestNode != nil {
			// 更新状态中的当前节点
			m.state.CurrentNode = mvpState.BestNode.Node
			m.state.LastUpdate = mvpState.LastUpdate

			// 构建ValidNodes列表
			m.state.ValidNodes = []types.ValidNode{*mvpState.BestNode}
		}
	}
}

// Stop 停止双进程自动代理系统
func (m *AutoProxyManager) Stop() error {
	fmt.Printf("🛑 停止双进程自动代理系统...\n")

	m.mutex.Lock()
	m.state.Running = false
	m.mutex.Unlock()

	// 停止测试进程
	fmt.Printf("  🛑 停止测试进程...\n")
	m.testerCancel()
	if m.tester != nil {
		m.tester.Stop()
	}

	// 停止代理服务进程
	fmt.Printf("  🛑 停止代理服务进程...\n")
	m.serverCancel()
	if m.proxyServer != nil {
		m.proxyServer.Stop()
	}

	// 停止主进程
	m.cancel()

	// 清理资源
	m.cleanup()

	// 保存最终状态
	m.saveState()

	fmt.Printf("✅ 双进程自动代理系统已停止\n")
	return nil
}

// GetStatus 获取系统状态
func (m *AutoProxyManager) GetStatus() types.AutoProxyState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 实时更新状态
	m.updateSystemStatus()
	return m.state
}

// GetBlacklistStatus 获取黑名单状态
func (m *AutoProxyManager) GetBlacklistStatus() map[string]time.Time {
	m.blacklistMutex.RLock()
	defer m.blacklistMutex.RUnlock()

	result := make(map[string]time.Time)
	for key, expireTime := range m.blacklist {
		result[key] = expireTime
	}
	return result
}

// 保留一些通用工具函数用于兼容性

// validateConfig 验证配置
func (m *AutoProxyManager) validateConfig() error {
	if m.config.SubscriptionURL == "" {
		return fmt.Errorf("订阅链接不能为空")
	}

	// 验证URL格式
	if _, err := url.Parse(m.config.SubscriptionURL); err != nil {
		return fmt.Errorf("订阅链接格式无效: %v", err)
	}

	// 验证端口范围
	if m.config.HTTPPort < 1024 || m.config.HTTPPort > 65535 {
		return fmt.Errorf("HTTP端口范围无效: %d", m.config.HTTPPort)
	}

	if m.config.SOCKSPort < 1024 || m.config.SOCKSPort > 65535 {
		return fmt.Errorf("SOCKS端口范围无效: %d", m.config.SOCKSPort)
	}

	// 验证时间间隔
	if m.config.UpdateInterval < time.Minute {
		return fmt.Errorf("更新间隔不能少于1分钟")
	}

	// 验证并发数
	if m.config.TestConcurrency < 1 || m.config.TestConcurrency > 100 {
		return fmt.Errorf("测试并发数范围无效: %d", m.config.TestConcurrency)
	}

	// 验证超时时间
	if m.config.TestTimeout < 5*time.Second {
		return fmt.Errorf("测试超时时间不能少于5秒")
	}

	return nil
}

// loadState 加载状态
func (m *AutoProxyManager) loadState() {
	if data, err := os.ReadFile(m.config.StateFile); err == nil {
		json.Unmarshal(data, &m.state)
	}
}

// saveState 保存状态
func (m *AutoProxyManager) saveState() {
	data, _ := json.MarshalIndent(m.state, "", "  ")
	os.WriteFile(m.config.StateFile, data, 0644)
}

// setupSignalHandler 设置信号处理
func (m *AutoProxyManager) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("\n🛑 接收到退出信号，正在停止双进程自动代理系统...\n")
		m.Stop()
		os.Exit(0)
	}()
}

// checkDependencies 检查依赖
func (m *AutoProxyManager) checkDependencies() error {
	fmt.Printf("🔧 检查系统依赖...\n")

	// 检查V2Ray
	v2rayDownloader := downloader.NewV2RayDownloader()
	if !v2rayDownloader.CheckV2rayInstalled() {
		fmt.Printf("📥 V2Ray未安装，正在自动下载...\n")
		if err := downloader.AutoDownloadV2Ray(); err != nil {
			return fmt.Errorf("V2Ray下载失败: %v", err)
		}
		fmt.Printf("✅ V2Ray安装完成\n")
	} else {
		fmt.Printf("✅ V2Ray已安装\n")
	}

	// 检查Hysteria2
	hysteria2Downloader := downloader.NewHysteria2Downloader()
	if !hysteria2Downloader.CheckHysteria2Installed() {
		fmt.Printf("📥 Hysteria2未安装，正在自动下载...\n")
		if err := downloader.AutoDownloadHysteria2(); err != nil {
			return fmt.Errorf("Hysteria2下载失败: %v", err)
		}
		fmt.Printf("✅ Hysteria2安装完成\n")
	} else {
		fmt.Printf("✅ Hysteria2已安装\n")
	}

	// 创建必要的目录
	dirs := []string{"./hysteria2", "./v2ray"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("⚠️ 创建目录失败: %s - %v\n", dir, err)
		}
	}

	fmt.Printf("✅ 所有依赖检查完成\n")
	return nil
}

// cleanup 清理系统资源
func (m *AutoProxyManager) cleanup() {
	fmt.Printf("🧹 清理系统资源...\n")

	// 清理过期黑名单
	m.cleanExpiredBlacklist()

	// 使用通用清理函数
	utils.ForceCleanupAll()

	// 杀死相关进程
	m.killRelatedProcesses()

	fmt.Printf("✅ 资源清理完成\n")
}

// cleanExpiredBlacklist 清理过期黑名单
func (m *AutoProxyManager) cleanExpiredBlacklist() {
	m.blacklistMutex.Lock()
	defer m.blacklistMutex.Unlock()

	now := time.Now()
	for key, expireTime := range m.blacklist {
		if now.After(expireTime) {
			delete(m.blacklist, key)
		}
	}
}

// killRelatedProcesses 杀死相关进程
func (m *AutoProxyManager) killRelatedProcesses() {
	processNames := []string{"v2ray", "xray", "hysteria2", "hysteria"}

	for _, processName := range processNames {
		cmd := exec.Command("pkill", "-f", processName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("    💀 已终止 %s 进程\n", processName)
		}
	}
}

// RunAutoProxy 运行双进程自动代理系统
func RunAutoProxy(config types.AutoProxyConfig) error {
	// 验证配置
	manager := NewAutoProxyManager(config)
	if err := manager.validateConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 启动系统
	if err := manager.Start(); err != nil {
		return fmt.Errorf("启动双进程自动代理系统失败: %v", err)
	}

	// 阻塞等待
	select {
	case <-manager.ctx.Done():
		return nil
	}
}
