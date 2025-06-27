package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/core/downloader"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/parser"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// AutoProxyManager 自动代理管理器
type AutoProxyManager struct {
	config           types.AutoProxyConfig
	state            types.AutoProxyState
	proxyManager     *proxy.ProxyManager
	hysteria2Manager *proxy.Hysteria2ProxyManager
	ctx              context.Context
	cancel           context.CancelFunc
	mutex            sync.RWMutex
	testMutex        sync.Mutex
	updateTicker     *time.Ticker
	testResults      []types.ValidNode
	currentProxy     interface{}          // 当前代理管理器实例
	blacklist        map[string]time.Time // 节点黑名单，key为节点标识，value为解禁时间
	blacklistMutex   sync.RWMutex         // 黑名单读写锁
}

// NewAutoProxyManager 创建新的自动代理管理器
func NewAutoProxyManager(config types.AutoProxyConfig) *AutoProxyManager {
	ctx, cancel := context.WithCancel(context.Background())

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

	return &AutoProxyManager{
		config:           config,
		proxyManager:     proxy.NewProxyManager(),
		hysteria2Manager: proxy.NewHysteria2ProxyManager(),
		ctx:              ctx,
		cancel:           cancel,
		testResults:      make([]types.ValidNode, 0),
		state: types.AutoProxyState{
			Config:     config,
			ValidNodes: make([]types.ValidNode, 0),
			StartTime:  time.Now(),
		},
		blacklist:      make(map[string]time.Time),
		blacklistMutex: sync.RWMutex{},
	}
}

// Start 启动自动代理系统
func (m *AutoProxyManager) Start() error {
	fmt.Printf("🚀 启动自动代理系统...\n")

	// 验证配置
	if err := m.validateConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	fmt.Printf("📡 订阅链接: %s\n", m.config.SubscriptionURL)
	fmt.Printf("🌐 HTTP代理: http://127.0.0.1:%d\n", m.config.HTTPPort)
	fmt.Printf("🧦 SOCKS代理: socks5://127.0.0.1:%d\n", m.config.SOCKSPort)
	fmt.Printf("⏰ 更新间隔: %v\n", m.config.UpdateInterval)

	// 设置信号处理
	m.setupSignalHandler()

	// 检查依赖
	if err := m.checkDependencies(); err != nil {
		return fmt.Errorf("依赖检查失败: %v", err)
	}

	// 加载历史状态
	m.loadState()

	// 启动后台任务
	m.state.Running = true
	m.state.StartTime = time.Now()

	// 启动测试任务协程
	go m.testWorker()

	// 启动代理更新协程
	go m.proxyUpdateWorker()

	// 初始化测试
	fmt.Printf("🧪 执行初始节点测试...\n")
	if err := m.performInitialTest(); err != nil {
		fmt.Printf("⚠️ 初始测试失败: %v\n", err)
	}

	// 启动定时更新
	m.updateTicker = time.NewTicker(m.config.UpdateInterval)
	go m.scheduledUpdateWorker()

	fmt.Printf("✅ 自动代理系统启动成功！\n")
	return nil
}

// Stop 停止自动代理系统
func (m *AutoProxyManager) Stop() error {
	fmt.Printf("🛑 停止自动代理系统...\n")

	m.mutex.Lock()
	m.state.Running = false
	m.mutex.Unlock()

	// 取消上下文
	m.cancel()

	// 停止定时器
	if m.updateTicker != nil {
		m.updateTicker.Stop()
	}

	// 停止代理
	if m.proxyManager != nil {
		m.proxyManager.StopProxy()
	}
	if m.hysteria2Manager != nil {
		m.hysteria2Manager.StopHysteria2Proxy()
	}

	// 清理资源
	m.cleanup()

	// 保存状态
	m.saveState()

	fmt.Printf("✅ 自动代理系统已停止\n")
	return nil
}

// GetStatus 获取系统状态
func (m *AutoProxyManager) GetStatus() types.AutoProxyState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state
}

// performInitialTest 执行初始测试
func (m *AutoProxyManager) performInitialTest() error {
	// 获取订阅内容
	nodes, err := m.fetchAndParseSubscription()
	if err != nil {
		return fmt.Errorf("获取订阅失败: %v", err)
	}

	fmt.Printf("📥 获取到 %d 个节点\n", len(nodes))

	// 限制测试节点数量
	if m.config.MaxNodes > 0 && len(nodes) > m.config.MaxNodes {
		nodes = nodes[:m.config.MaxNodes]
		fmt.Printf("🔢 限制测试节点数量为: %d\n", len(nodes))
	}

	// 批量测试节点
	validNodes := m.batchTestNodes(nodes)
	fmt.Printf("✅ 测试完成，发现 %d 个有效节点\n", len(validNodes))

	// 更新状态
	m.mutex.Lock()
	m.state.ValidNodes = validNodes
	m.state.LastUpdate = time.Now()
	m.mutex.Unlock()

	// 保存有效节点到文件
	m.saveValidNodesToFile(validNodes)

	// 如果有有效节点，启动最优代理
	if len(validNodes) > 0 {
		return m.switchToBestNode()
	}

	return fmt.Errorf("没有发现有效节点")
}

// fetchAndParseSubscription 获取并解析订阅
func (m *AutoProxyManager) fetchAndParseSubscription() ([]*types.Node, error) {
	// 获取订阅内容
	content, err := parser.FetchSubscription(m.config.SubscriptionURL)
	if err != nil {
		return nil, err
	}

	// 解码base64
	decoded, err := parser.DecodeBase64(content)
	if err != nil {
		return nil, err
	}

	// 解析所有链接
	nodes, err := parser.ParseLinks(decoded)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// batchTestNodes 批量测试节点
func (m *AutoProxyManager) batchTestNodes(nodes []*types.Node) []types.ValidNode {
	fmt.Printf("🧪 开始批量测试 %d 个节点（并发数: %d）...\n", len(nodes), m.config.TestConcurrency)

	var wg sync.WaitGroup
	nodeQueue := make(chan *types.Node, len(nodes))
	resultQueue := make(chan types.ValidNode, len(nodes))

	// 发送节点到队列
	for _, node := range nodes {
		nodeQueue <- node
	}
	close(nodeQueue)

	// 启动工作协程
	for i := 0; i < m.config.TestConcurrency; i++ {
		wg.Add(1)
		go m.nodeTestWorker(nodeQueue, resultQueue, &wg, i)
	}

	// 等待所有测试完成
	go func() {
		wg.Wait()
		close(resultQueue)
	}()

	// 收集结果
	var validNodes []types.ValidNode
	for result := range resultQueue {
		if result.Node != nil {
			validNodes = append(validNodes, result)
		}
	}

	// 按评分排序
	sort.Slice(validNodes, func(i, j int) bool {
		return validNodes[i].Score > validNodes[j].Score
	})

	return validNodes
}

// nodeTestWorker 节点测试工作协程
func (m *AutoProxyManager) nodeTestWorker(nodeQueue <-chan *types.Node, resultQueue chan<- types.ValidNode, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	portBase := 9000 + workerID*10 // 为每个worker分配不同的端口范围

	for node := range nodeQueue {
		select {
		case <-m.ctx.Done():
			return
		default:
			// 检查节点是否在黑名单中
			if m.isNodeBlacklisted(node) {
				fmt.Printf("🚫 [%s] 节点在黑名单中，跳过测试\n", node.Name)
				continue
			}

			result := m.testSingleNodeWithScore(node, portBase)
			if result.Node != nil {
				resultQueue <- result
				// 更新性能历史
				m.updateNodePerformanceHistory(result)
			}
		}
	}
}

// testSingleNodeWithScore 测试单个节点并计算评分
func (m *AutoProxyManager) testSingleNodeWithScore(node *types.Node, portBase int) types.ValidNode {
	result := types.ValidNode{
		TestTime: time.Now(),
	}

	// 测试节点，支持重试
	var latency int64
	var speed float64
	var err error

	overallStart := time.Now()
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		latency, speed, err = m.testNodeConnectivity(node, portBase)
		if err == nil {
			break // 成功则跳出重试循环
		}

		// 处理测试错误
		m.handleNodeTestError(node, err, retry+1)

		// 如果不是最后一次重试，等待一段时间再重试
		if retry < maxRetries-1 {
			time.Sleep(time.Duration(retry+1) * 2 * time.Second)
		}
	}

	if err != nil {
		fmt.Printf("❌ [%s] 测试最终失败: %v\n", node.Name, err)
		return result
	}

	result.Node = node
	result.Latency = latency
	result.Speed = speed
	result.SuccessCount = 1

	// 计算综合评分 (延迟权重0.3, 速度权重0.7)
	latencyScore := math.Max(0, 100-float64(latency)/10) // 延迟越低分数越高
	speedScore := math.Min(100, speed*10)                // 速度越高分数越高
	result.Score = latencyScore*0.3 + speedScore*0.7

	elapsed := time.Since(overallStart)
	fmt.Printf("✅ [%s] 延迟:%dms 速度:%.2fMbps 评分:%.1f 耗时:%v\n",
		node.Name, latency, speed, result.Score, elapsed)

	return result
}

// testNodeConnectivity 测试节点连通性
func (m *AutoProxyManager) testNodeConnectivity(node *types.Node, portBase int) (int64, float64, error) {
	var httpPort int

	// 根据协议创建对应的代理管理器
	switch node.Protocol {
	case "hysteria2":
		hysteria2Mgr := proxy.NewHysteria2ProxyManager()
		hysteria2Mgr.HTTPPort = portBase + 1
		hysteria2Mgr.SOCKSPort = portBase + 2
		httpPort = hysteria2Mgr.HTTPPort

		if err := hysteria2Mgr.StartHysteria2Proxy(node); err != nil {
			return 0, 0, fmt.Errorf("启动Hysteria2代理失败: %v", err)
		}

		defer hysteria2Mgr.StopHysteria2Proxy()

	default:
		v2rayMgr := proxy.NewProxyManager()
		v2rayMgr.HTTPPort = portBase + 3
		v2rayMgr.SOCKSPort = portBase + 4
		httpPort = v2rayMgr.HTTPPort

		if err := v2rayMgr.StartProxy(node); err != nil {
			return 0, 0, fmt.Errorf("启动V2Ray代理失败: %v", err)
		}

		defer v2rayMgr.StopProxy()
	}

	// 等待代理启动
	time.Sleep(2 * time.Second)

	// 测试连通性和速度
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	return m.measureProxyPerformance(proxyURL)
}

// measureProxyPerformance 测量代理性能
func (m *AutoProxyManager) measureProxyPerformance(proxyURL string) (int64, float64, error) {
	// 创建代理客户端
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		return 0, 0, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURLParsed),
		},
		Timeout: m.config.TestTimeout,
	}

	// 测试延迟
	start := time.Now()
	resp, err := client.Get(m.config.TestURL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return 0, 0, fmt.Errorf("连接测试失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 简单的速度测试（读取响应数据）
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return latency, 0, nil // 只返回延迟，速度为0
	}

	// 估算速度（简化计算）
	downloadTime := time.Since(start).Seconds()
	if downloadTime > 0 {
		speed := float64(len(data)) / downloadTime / 1024 / 1024 * 8 // Mbps
		return latency, speed, nil
	}

	return latency, 0, nil
}

// switchToBestNode 切换到最佳节点
func (m *AutoProxyManager) switchToBestNode() error {
	m.mutex.RLock()
	validNodes := m.state.ValidNodes
	m.mutex.RUnlock()

	if len(validNodes) == 0 {
		return fmt.Errorf("没有有效节点")
	}

	bestNode := validNodes[0] // 已经按评分排序
	fmt.Printf("🔄 切换到最佳节点: %s (评分: %.1f)\n", bestNode.Node.Name, bestNode.Score)

	// 停止当前代理
	if m.proxyManager != nil {
		m.proxyManager.StopProxy()
	}
	if m.hysteria2Manager != nil {
		m.hysteria2Manager.StopHysteria2Proxy()
	}

	// 启动新代理
	switch bestNode.Node.Protocol {
	case "hysteria2":
		m.hysteria2Manager.HTTPPort = m.config.HTTPPort
		m.hysteria2Manager.SOCKSPort = m.config.SOCKSPort
		if err := m.hysteria2Manager.StartHysteria2Proxy(bestNode.Node); err != nil {
			return fmt.Errorf("启动Hysteria2代理失败: %v", err)
		}
		m.currentProxy = m.hysteria2Manager

	default:
		m.proxyManager.HTTPPort = m.config.HTTPPort
		m.proxyManager.SOCKSPort = m.config.SOCKSPort
		if err := m.proxyManager.StartProxy(bestNode.Node); err != nil {
			return fmt.Errorf("启动V2Ray代理失败: %v", err)
		}
		m.currentProxy = m.proxyManager
	}

	// 更新状态
	m.mutex.Lock()
	m.state.CurrentNode = bestNode.Node
	m.state.LastUpdate = time.Now()
	m.mutex.Unlock()

	fmt.Printf("✅ 代理切换完成！\n")
	fmt.Printf("🌐 HTTP代理: http://127.0.0.1:%d\n", m.config.HTTPPort)
	fmt.Printf("🧦 SOCKS代理: socks5://127.0.0.1:%d\n", m.config.SOCKSPort)

	return nil
}

// testWorker 测试工作协程
func (m *AutoProxyManager) testWorker() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// 监控当前代理状态
			m.monitorCurrentProxy()

			// 每5分钟清理一次过期黑名单
			if time.Now().Minute()%5 == 0 {
				m.cleanExpiredBlacklist()
			}
		}
	}
}

// proxyUpdateWorker 代理更新工作协程
func (m *AutoProxyManager) proxyUpdateWorker() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(5 * time.Minute): // 每5分钟检查一次是否需要切换
			if m.config.EnableAutoSwitch {
				m.checkAndSwitchProxy()
			}
		}
	}
}

// scheduledUpdateWorker 定时更新工作协程
func (m *AutoProxyManager) scheduledUpdateWorker() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.updateTicker.C:
			fmt.Printf("⏰ 开始定时更新节点测试...\n")
			if err := m.performScheduledUpdate(); err != nil {
				fmt.Printf("❌ 定时更新失败: %v\n", err)
			}
		}
	}
}

// performScheduledUpdate 执行定时更新
func (m *AutoProxyManager) performScheduledUpdate() error {
	m.testMutex.Lock()
	defer m.testMutex.Unlock()

	// 获取最新订阅
	nodes, err := m.fetchAndParseSubscription()
	if err != nil {
		return err
	}

	// 限制节点数量
	if m.config.MaxNodes > 0 && len(nodes) > m.config.MaxNodes {
		nodes = nodes[:m.config.MaxNodes]
	}

	// 批量测试
	validNodes := m.batchTestNodes(nodes)

	// 更新状态
	m.mutex.Lock()
	m.state.ValidNodes = validNodes
	m.state.LastUpdate = time.Now()
	m.state.TotalTests++
	if len(validNodes) > 0 {
		m.state.SuccessfulTests++
	}
	m.mutex.Unlock()

	// 保存到文件
	m.saveValidNodesToFile(validNodes)

	fmt.Printf("✅ 定时更新完成，发现 %d 个有效节点\n", len(validNodes))

	// 如果启用自动切换，切换到最佳节点
	if m.config.EnableAutoSwitch && len(validNodes) > 0 {
		return m.switchToBestNode()
	}

	return nil
}

// monitorCurrentProxy 监控当前代理
func (m *AutoProxyManager) monitorCurrentProxy() {
	if m.state.CurrentNode == nil {
		return
	}

	// 测试当前代理是否正常
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", m.config.HTTPPort)
	_, _, err := m.measureProxyPerformance(proxyURL)

	if err != nil {
		fmt.Printf("⚠️ 当前代理异常: %v\n", err)
		if m.config.EnableAutoSwitch {
			m.checkAndSwitchProxy()
		}
	}
}

// checkAndSwitchProxy 检查并切换代理
func (m *AutoProxyManager) checkAndSwitchProxy() {
	m.mutex.RLock()
	validNodes := m.state.ValidNodes
	currentNode := m.state.CurrentNode
	m.mutex.RUnlock()

	if len(validNodes) == 0 {
		return
	}

	// 如果当前节点不是最佳节点，则切换
	if currentNode == nil ||
		len(validNodes) > 0 &&
			(currentNode.Name != validNodes[0].Node.Name || currentNode.Server != validNodes[0].Node.Server) {

		fmt.Printf("🔄 检测到更好的节点，准备切换...\n")
		if err := m.switchToBestNode(); err != nil {
			fmt.Printf("❌ 自动切换失败: %v\n", err)
		}
	}
}

// saveValidNodesToFile 保存有效节点到文件
func (m *AutoProxyManager) saveValidNodesToFile(validNodes []types.ValidNode) error {
	data, err := json.MarshalIndent(validNodes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.config.ValidNodesFile, data, 0644)
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
		fmt.Printf("\n🛑 接收到退出信号，正在停止自动代理系统...\n")
		m.Stop()
		os.Exit(0)
	}()
}

// checkDependencies 检查依赖
func (m *AutoProxyManager) checkDependencies() error {
	// 检查V2Ray
	v2rayDownloader := downloader.NewV2RayDownloader()
	if !v2rayDownloader.CheckV2rayInstalled() {
		fmt.Printf("📥 V2Ray未安装，正在自动下载...\n")
		if err := downloader.AutoDownloadV2Ray(); err != nil {
			return fmt.Errorf("V2Ray下载失败: %v", err)
		}
	}

	// 检查Hysteria2
	hysteria2Downloader := downloader.NewHysteria2Downloader()
	if !hysteria2Downloader.CheckHysteria2Installed() {
		fmt.Printf("📥 Hysteria2未安装，正在自动下载...\n")
		if err := downloader.AutoDownloadHysteria2(); err != nil {
			return fmt.Errorf("Hysteria2下载失败: %v", err)
		}
	}

	return nil
}

// 节点黑名单管理

// getNodeKey 获取节点唯一标识
func (m *AutoProxyManager) getNodeKey(node *types.Node) string {
	return fmt.Sprintf("%s:%s:%d", node.Protocol, node.Server, node.Port)
}

// isNodeBlacklisted 检查节点是否在黑名单中
func (m *AutoProxyManager) isNodeBlacklisted(node *types.Node) bool {
	m.blacklistMutex.RLock()
	defer m.blacklistMutex.RUnlock()

	key := m.getNodeKey(node)
	if expireTime, exists := m.blacklist[key]; exists {
		if time.Now().Before(expireTime) {
			return true
		}
		// 如果已过期，从黑名单中移除
		delete(m.blacklist, key)
	}
	return false
}

// addToBlacklist 将节点添加到黑名单
func (m *AutoProxyManager) addToBlacklist(node *types.Node, duration time.Duration) {
	m.blacklistMutex.Lock()
	defer m.blacklistMutex.Unlock()

	key := m.getNodeKey(node)
	expireTime := time.Now().Add(duration)
	m.blacklist[key] = expireTime

	fmt.Printf("🚫 节点 [%s] 已加入黑名单，解禁时间: %s\n",
		node.Name, expireTime.Format("15:04:05"))
}

// removeFromBlacklist 从黑名单中移除节点
func (m *AutoProxyManager) removeFromBlacklist(node *types.Node) {
	m.blacklistMutex.Lock()
	defer m.blacklistMutex.Unlock()

	key := m.getNodeKey(node)
	delete(m.blacklist, key)
}

// cleanExpiredBlacklist 清理过期的黑名单条目
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

// 配置验证

// validateConfig 验证配置有效性
func (m *AutoProxyManager) validateConfig() error {
	config := m.config

	// 验证订阅URL
	if config.SubscriptionURL == "" {
		return fmt.Errorf("订阅URL不能为空")
	}

	if _, err := url.Parse(config.SubscriptionURL); err != nil {
		return fmt.Errorf("无效的订阅URL: %v", err)
	}

	// 验证端口
	if config.HTTPPort <= 0 || config.HTTPPort > 65535 {
		return fmt.Errorf("无效的HTTP端口: %d", config.HTTPPort)
	}

	if config.SOCKSPort <= 0 || config.SOCKSPort > 65535 {
		return fmt.Errorf("无效的SOCKS端口: %d", config.SOCKSPort)
	}

	if config.HTTPPort == config.SOCKSPort {
		return fmt.Errorf("HTTP端口和SOCKS端口不能相同")
	}

	// 验证时间间隔
	if config.UpdateInterval < time.Minute {
		return fmt.Errorf("更新间隔不能少于1分钟")
	}

	if config.TestTimeout < 5*time.Second {
		return fmt.Errorf("测试超时不能少于5秒")
	}

	// 验证并发数
	if config.TestConcurrency <= 0 || config.TestConcurrency > 100 {
		return fmt.Errorf("测试并发数必须在1-100之间")
	}

	// 验证测试URL
	if config.TestURL == "" {
		return fmt.Errorf("测试URL不能为空")
	}

	if _, err := url.Parse(config.TestURL); err != nil {
		return fmt.Errorf("无效的测试URL: %v", err)
	}

	// 验证节点数量限制
	if config.MaxNodes < 0 {
		return fmt.Errorf("最大节点数不能为负数")
	}

	if config.MinPassingNodes < 0 {
		return fmt.Errorf("最少通过节点数不能为负数")
	}

	return nil
}

// 增强的错误处理

// handleNodeTestError 处理节点测试错误
func (m *AutoProxyManager) handleNodeTestError(node *types.Node, err error, retryCount int) {
	fmt.Printf("❌ [%s] 测试失败 (重试:%d): %v\n", node.Name, retryCount, err)

	// 如果重试次数过多，加入黑名单
	if retryCount >= 3 {
		// 根据错误类型决定黑名单时长
		var blacklistDuration time.Duration
		errStr := err.Error()

		if strings.Contains(errStr, "timeout") {
			blacklistDuration = 30 * time.Minute // 超时错误，30分钟黑名单
		} else if strings.Contains(errStr, "connection refused") {
			blacklistDuration = 1 * time.Hour // 连接拒绝，1小时黑名单
		} else if strings.Contains(errStr, "no route to host") {
			blacklistDuration = 2 * time.Hour // 无路由，2小时黑名单
		} else {
			blacklistDuration = 15 * time.Minute // 其他错误，15分钟黑名单
		}

		m.addToBlacklist(node, blacklistDuration)
	}
}

// logError 记录错误到状态
func (m *AutoProxyManager) logError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.LastError = err.Error()
	fmt.Printf("⚠️ 系统错误: %v\n", err)
}

// 性能监控

// updateNodePerformanceHistory 更新节点性能历史
func (m *AutoProxyManager) updateNodePerformanceHistory(validNode types.ValidNode) {
	// 这里可以实现节点性能历史记录功能
	// 为了简化，暂时只记录到内存中
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 查找现有节点记录并更新
	for i, existing := range m.state.ValidNodes {
		if existing.Node != nil && validNode.Node != nil &&
			m.getNodeKey(existing.Node) == m.getNodeKey(validNode.Node) {

			// 更新成功/失败计数
			m.state.ValidNodes[i].SuccessCount = existing.SuccessCount + 1
			m.state.ValidNodes[i].TestTime = validNode.TestTime
			m.state.ValidNodes[i].Latency = validNode.Latency
			m.state.ValidNodes[i].Speed = validNode.Speed
			m.state.ValidNodes[i].Score = validNode.Score
			return
		}
	}
}

// 资源清理

// cleanup 清理资源
func (m *AutoProxyManager) cleanup() {
	fmt.Printf("🧹 清理系统资源...\n")

	// 清理过期黑名单
	m.cleanExpiredBlacklist()

	// 清理临时文件（如果有的话）
	// 这里可以添加清理临时配置文件的逻辑

	fmt.Printf("✅ 资源清理完成\n")
}
