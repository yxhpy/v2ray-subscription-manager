package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/core/downloader"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/parser"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// MVPTester MVP节点测试器
type MVPTester struct {
	subscriptionURL  string
	bestNode         *types.ValidNode
	mutex            sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	testInterval     time.Duration
	stateFile        string
	maxNodes         int
	concurrency      int
	proxyManager     *proxy.ProxyManager
	hysteria2Manager *proxy.Hysteria2ProxyManager

	// 添加配置字段
	testTimeout time.Duration
	testURL     string
}

// MVPState MVP状态
type MVPState struct {
	BestNode   *types.ValidNode `json:"best_node"`
	LastUpdate time.Time        `json:"last_update"`
	TestCount  int              `json:"test_count"`
	TotalNodes int              `json:"total_nodes"`
	ValidNodes int              `json:"valid_nodes"`
}

// NewMVPTester 创建新的MVP测试器
func NewMVPTester(subscriptionURL string) *MVPTester {
	ctx, cancel := context.WithCancel(context.Background())

	// 根据平台设置默认超时时间
	defaultTimeout := 30 * time.Second
	defaultTestURL := "http://www.google.com"
	if runtime.GOOS == "windows" {
		defaultTimeout = 60 * time.Second       // Windows下使用更长超时
		defaultTestURL = "http://www.baidu.com" // Windows下使用百度
	}

	return &MVPTester{
		subscriptionURL:  subscriptionURL,
		ctx:              ctx,
		cancel:           cancel,
		testInterval:     5 * time.Minute, // 每5分钟测试一次
		stateFile:        "mvp_best_node.json",
		maxNodes:         50,
		concurrency:      5,
		proxyManager:     proxy.NewProxyManager(),
		hysteria2Manager: proxy.NewHysteria2ProxyManager(),

		// 使用平台相关的默认值
		testTimeout: defaultTimeout,
		testURL:     defaultTestURL,
	}
}

// SetInterval 设置测试间隔
func (m *MVPTester) SetInterval(interval time.Duration) {
	m.testInterval = interval
}

// SetMaxNodes 设置最大测试节点数
func (m *MVPTester) SetMaxNodes(maxNodes int) {
	m.maxNodes = maxNodes
}

// SetConcurrency 设置测试并发数
func (m *MVPTester) SetConcurrency(concurrency int) {
	m.concurrency = concurrency
}

// SetStateFile 设置状态文件路径
func (m *MVPTester) SetStateFile(stateFile string) {
	m.stateFile = stateFile
}

// SetTimeout 设置测试超时时间
func (m *MVPTester) SetTimeout(timeout time.Duration) {
	m.testTimeout = timeout
}

// SetTestURL 设置测试URL
func (m *MVPTester) SetTestURL(testURL string) {
	m.testURL = testURL
}

// Start 启动MVP测试器
func (m *MVPTester) Start() error {
	fmt.Printf("🚀 启动MVP节点测试器...\n")
	fmt.Printf("📡 订阅链接: %s\n", m.subscriptionURL)
	fmt.Printf("⏰ 测试间隔: %v\n", m.testInterval)
	fmt.Printf("💾 状态文件: %s\n", m.stateFile)

	// 设置信号处理
	m.setupSignalHandler()

	// 检查依赖
	if err := m.checkDependencies(); err != nil {
		return fmt.Errorf("依赖检查失败: %v", err)
	}

	// 加载历史最佳节点
	m.loadBestNode()

	// 立即执行一次测试
	fmt.Printf("🧪 执行初始节点测试...\n")
	if err := m.performTest(); err != nil {
		fmt.Printf("⚠️ 初始测试失败: %v\n", err)
	}

	// 启动定时测试
	ticker := time.NewTicker(m.testInterval)
	defer ticker.Stop()

	fmt.Printf("✅ MVP节点测试器启动成功！\n")
	fmt.Printf("📝 按 Ctrl+C 停止服务\n")

	for {
		select {
		case <-ticker.C:
			fmt.Printf("\n⏰ 开始定时测试 [%s]\n", time.Now().Format("2006-01-02 15:04:05"))
			if err := m.performTest(); err != nil {
				fmt.Printf("❌ 定时测试失败: %v\n", err)
			}
		case <-m.ctx.Done():
			fmt.Printf("\n🛑 收到停止信号，正在退出...\n")
			return nil
		}
	}
}

// Stop 停止MVP测试器
func (m *MVPTester) Stop() error {
	fmt.Printf("🛑 停止MVP测试器...\n")

	// 第一步：取消上下文
	m.cancel()

	// 第二步：停止V2Ray代理并等待
	if m.proxyManager != nil {
		fmt.Printf("  🛑 停止V2Ray代理...\n")
		if err := m.proxyManager.StopProxy(); err != nil {
			fmt.Printf("    ⚠️ V2Ray代理停止异常: %v\n", err)
		}
		m.waitForProxyStop("V2Ray", m.proxyManager)
	}

	// 第三步：停止Hysteria2代理并等待
	if m.hysteria2Manager != nil {
		fmt.Printf("  🛑 停止Hysteria2代理...\n")
		if err := m.hysteria2Manager.StopHysteria2Proxy(); err != nil {
			fmt.Printf("    ⚠️ Hysteria2代理停止异常: %v\n", err)
		}
		m.waitForHysteria2Stop("Hysteria2", m.hysteria2Manager)
	}

	// 第四步：等待所有操作完成
	fmt.Printf("  ⏳ 等待所有操作完成...\n")
	time.Sleep(3 * time.Second)

	// 第五步：强制终止残留进程
	fmt.Printf("  💀 强制终止残留进程...\n")
	m.killRelatedProcesses()

	// 第六步：等待进程终止完成
	time.Sleep(2 * time.Second)

	// 第七步：清理临时配置文件
	fmt.Printf("  🧹 清理临时配置文件...\n")
	m.cleanupTempFiles()

	// 第八步：清理状态文件
	fmt.Printf("  🧹 清理状态文件...\n")
	m.cleanupStateFile()

	// 第九步：验证清理结果
	m.verifyMVPCleanup()

	fmt.Printf("✅ MVP测试器已完全停止\n")
	return nil
}

// waitForProxyStop 等待V2Ray代理停止
func (m *MVPTester) waitForProxyStop(name string, manager *proxy.ProxyManager) {
	maxWait := 10 * time.Second
	interval := 500 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		if !manager.GetStatus().Running {
			fmt.Printf("    ✅ %s代理已停止\n", name)
			return
		}
		time.Sleep(interval)
		elapsed += interval
	}

	fmt.Printf("    ⚠️ %s代理停止超时\n", name)
}

// waitForHysteria2Stop 等待Hysteria2代理停止
func (m *MVPTester) waitForHysteria2Stop(name string, manager *proxy.Hysteria2ProxyManager) {
	maxWait := 10 * time.Second
	interval := 500 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		if !manager.GetHysteria2Status().Running {
			fmt.Printf("    ✅ %s代理已停止\n", name)
			return
		}
		time.Sleep(interval)
		elapsed += interval
	}

	fmt.Printf("    ⚠️ %s代理停止超时\n", name)
}

// verifyMVPCleanup 验证MVP清理结果
func (m *MVPTester) verifyMVPCleanup() {
	fmt.Printf("  🔍 验证MVP清理结果...\n")

	// 检查状态文件是否已删除
	if m.stateFile != "" {
		if _, err := os.Stat(m.stateFile); err == nil {
			fmt.Printf("    ⚠️ 状态文件仍存在: %s，尝试再次删除\n", m.stateFile)
			if err := os.Remove(m.stateFile); err != nil {
				fmt.Printf("    ❌ 删除失败: %s - %v\n", m.stateFile, err)
			} else {
				fmt.Printf("    ✅ 重试删除成功: %s\n", m.stateFile)
			}
		}
	}

	fmt.Printf("    ✅ MVP清理验证完成\n")
}

// cleanupStateFile 清理状态文件
func (m *MVPTester) cleanupStateFile() {
	if m.stateFile != "" {
		if err := os.Remove(m.stateFile); err == nil {
			fmt.Printf("    🗑️  已删除状态文件: %s\n", m.stateFile)
		}
	}
}

// cleanupTempFiles 清理临时文件
func (m *MVPTester) cleanupTempFiles() {
	patterns := []string{
		"temp_v2ray_config_*.json",
		"temp_hysteria2_config_*.json",
		"test_proxy_*.json", // 添加test_proxy_开头的文件
		"*.tmp",
		"*.temp",
	}

	cleanedCount := 0
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range matches {
			if err := os.Remove(file); err == nil {
				fmt.Printf("    🗑️  已删除: %s\n", file)
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		fmt.Printf("    ✅ 共清理了 %d 个临时文件\n", cleanedCount)
	}
}

// killRelatedProcesses 杀死相关进程
func (m *MVPTester) killRelatedProcesses() {
	fmt.Printf("    💀 终止MVP相关进程...\n")

	processNames := []string{"v2ray", "xray", "hysteria2", "hysteria"}

	if runtime.GOOS == "windows" {
		// Windows 使用taskkill
		for _, processName := range processNames {
			cmd := exec.Command("taskkill", "/F", "/IM", processName+".exe")
			if err := cmd.Run(); err == nil {
				fmt.Printf("      🔧 已终止 %s 进程\n", processName)
			}
		}
	} else {
		// Unix 使用pkill
		for _, processName := range processNames {
			cmd := exec.Command("pkill", "-f", processName)
			if err := cmd.Run(); err == nil {
				fmt.Printf("      🔧 已终止 %s 进程\n", processName)
			}
		}
	}
}

// performTest 执行测试
func (m *MVPTester) performTest() error {
	// 获取订阅内容
	nodes, err := m.fetchAndParseSubscription()
	if err != nil {
		return fmt.Errorf("获取订阅失败: %v", err)
	}

	fmt.Printf("📥 获取到 %d 个节点\n", len(nodes))

	if len(nodes) == 0 {
		return fmt.Errorf("没有找到任何节点")
	}

	// 测试所有节点
	validNodes := m.testAllNodes(nodes)
	fmt.Printf("✅ 测试完成，发现 %d 个有效节点\n", len(validNodes))

	if len(validNodes) == 0 {
		fmt.Printf("❌ 没有找到有效节点\n")
		return nil
	}

	// 按速度排序，找到最快的节点
	sort.Slice(validNodes, func(i, j int) bool {
		return validNodes[i].Score > validNodes[j].Score // 分数越高越好
	})

	newBestNode := &validNodes[0]

	// 检查是否需要更新最佳节点
	m.mutex.Lock()
	needUpdate := m.bestNode == nil || newBestNode.Score > m.bestNode.Score
	if needUpdate {
		oldBest := m.bestNode
		m.bestNode = newBestNode

		fmt.Printf("\n🎉 发现更快的节点！\n")
		if oldBest != nil {
			fmt.Printf("📊 旧节点: %s (分数: %.2f, 延迟: %dms, 速度: %.2fMbps)\n",
				oldBest.Node.Name, oldBest.Score, oldBest.Latency, oldBest.Speed)
		}
		fmt.Printf("🚀 新节点: %s (分数: %.2f, 延迟: %dms, 速度: %.2fMbps)\n",
			newBestNode.Node.Name, newBestNode.Score, newBestNode.Latency, newBestNode.Speed)

		// 保存到文件
		if err := m.saveBestNode(); err != nil {
			fmt.Printf("⚠️ 保存最佳节点失败: %v\n", err)
		} else {
			fmt.Printf("💾 最佳节点已保存到 %s\n", m.stateFile)
		}
	} else {
		fmt.Printf("📊 当前最佳节点仍是最快的: %s (分数: %.2f)\n",
			m.bestNode.Node.Name, m.bestNode.Score)
	}
	m.mutex.Unlock()

	// 显示测试摘要
	m.showTestSummary(validNodes)

	return nil
}

// fetchAndParseSubscription 获取并解析订阅
func (m *MVPTester) fetchAndParseSubscription() ([]*types.Node, error) {
	// 获取订阅内容
	content, err := parser.FetchSubscription(m.subscriptionURL)
	if err != nil {
		return nil, fmt.Errorf("获取订阅内容失败: %v", err)
	}

	// 解码base64（如果需要）
	decodedContent, err := parser.DecodeBase64(content)
	if err != nil {
		return nil, fmt.Errorf("解码订阅内容失败: %v", err)
	}

	// 解析节点
	nodes, err := parser.ParseLinks(decodedContent)
	if err != nil {
		return nil, fmt.Errorf("解析节点失败: %v", err)
	}

	return nodes, nil
}

// testAllNodes 测试所有节点
func (m *MVPTester) testAllNodes(nodes []*types.Node) []types.ValidNode {
	var validNodes []types.ValidNode
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// 使用用户通过SetConcurrency设置的并发数
	concurrency := m.concurrency

	// 如果并发数为0或过大，则使用平台相关的默认值作为安全后备
	if concurrency <= 0 {
		if runtime.GOOS == "windows" {
			concurrency = 1 // Windows下默认单线程
		} else {
			concurrency = 2 // Unix环境默认2个并发
		}
		fmt.Printf("⚠️ 并发数未设置或无效，使用默认值: %d\n", concurrency)
	} else {
		fmt.Printf("🔧 使用设置的并发数: %d\n", concurrency)
	}

	// Windows环境提示
	if runtime.GOOS == "windows" {
		fmt.Printf("🪟 Windows环境：并发数 = %d\n", concurrency)
	}

	semaphore := make(chan struct{}, concurrency)

	// 添加总体超时控制
	totalTimeout := 30 * time.Minute // 总测试时间限制
	if runtime.GOOS == "windows" {
		totalTimeout = 45 * time.Minute // Windows下允许更长时间
	}

	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	// 添加快速跳过机制
	var consecutiveFailures int
	var failureMutex sync.Mutex
	maxConsecutiveFailures := 10 // 连续失败10个节点后，缩短测试时间

	for i, node := range nodes {
		// 检查是否超时
		select {
		case <-ctx.Done():
			fmt.Printf("⏰ 总体测试超时，停止后续节点测试\n")
			break
		default:
		}

		// 检查是否应该快速跳过
		failureMutex.Lock()
		shouldFastFail := consecutiveFailures >= maxConsecutiveFailures
		failureMutex.Unlock()

		if shouldFastFail && runtime.GOOS == "windows" {
			fmt.Printf("⚡ 连续失败过多，启用快速测试模式\n")
		}

		wg.Add(1)
		go func(node *types.Node, index int, fastFail bool) {
			defer wg.Done()

			// 获取信号量
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				fmt.Printf("❌ 节点 [%d/%d] %s: 测试超时取消\n", index+1, len(nodes), node.Name)
				return
			}

			fmt.Printf("🧪 测试节点 [%d/%d]: %s (%s)\n",
				index+1, len(nodes), node.Name, node.Protocol)

			// 为单个节点测试添加超时，快速失败模式下缩短时间
			nodeTimeout := 3 * time.Minute
			if runtime.GOOS == "windows" {
				if fastFail {
					nodeTimeout = 1 * time.Minute // 快速失败模式
				} else {
					nodeTimeout = 5 * time.Minute // 正常模式
				}
			}

			nodeCtx, nodeCancel := context.WithTimeout(ctx, nodeTimeout)
			defer nodeCancel()

			// 在goroutine中执行测试，以便可以被取消
			resultChan := make(chan types.ValidNode, 1)
			go func() {
				result := m.testSingleNode(node, 8000+index*10)
				select {
				case resultChan <- result:
				case <-nodeCtx.Done():
				}
			}()

			var validNode types.ValidNode
			select {
			case validNode = <-resultChan:
				// 测试完成
			case <-nodeCtx.Done():
				fmt.Printf("⏰ 节点 [%d/%d] %s: 单节点测试超时\n", index+1, len(nodes), node.Name)
				// 记录失败
				failureMutex.Lock()
				consecutiveFailures++
				failureMutex.Unlock()
				return
			}

			if validNode.Node != nil {
				// 成功，重置连续失败计数
				failureMutex.Lock()
				consecutiveFailures = 0
				failureMutex.Unlock()

				mutex.Lock()
				validNodes = append(validNodes, validNode)

				// 检查是否是更好的节点
				if m.bestNode == nil || validNode.Score > m.bestNode.Score {
					m.bestNode = &validNode
					fmt.Printf("🏆 发现新的最佳节点: %s (分数: %.2f)\n", validNode.Node.Name, validNode.Score)

					// 立即保存最佳节点
					if err := m.saveBestNode(); err != nil {
						fmt.Printf("⚠️ 保存最佳节点失败: %v\n", err)
					} else {
						fmt.Printf("💾 最佳节点已保存到文件\n")
					}
				}
				mutex.Unlock()

				fmt.Printf("✅ 节点 %s 测试通过 (延迟: %dms, 速度: %.2fMbps, 分数: %.2f)\n",
					node.Name, validNode.Latency, validNode.Speed, validNode.Score)
			} else {
				// 失败，增加连续失败计数
				failureMutex.Lock()
				consecutiveFailures++
				failureMutex.Unlock()

				fmt.Printf("❌ 节点 %s 测试失败\n", node.Name)
			}
		}(node, i, shouldFastFail)

		// Windows环境在节点之间添加短暂延迟，但快速失败模式下减少延迟
		if runtime.GOOS == "windows" && concurrency == 1 {
			if shouldFastFail {
				time.Sleep(500 * time.Millisecond) // 快速模式
			} else {
				time.Sleep(2 * time.Second) // 正常模式
			}
		}
	}

	wg.Wait()
	return validNodes
}

// testSingleNode 测试单个节点
func (m *MVPTester) testSingleNode(node *types.Node, portBase int) types.ValidNode {
	result := types.ValidNode{
		TestTime: time.Now(),
	}

	switch node.Protocol {
	case "vmess", "vless", "trojan", "ss":
		result = m.testV2RayNode(node, result, portBase)
	case "hysteria2":
		result = m.testHysteria2Node(node, result, portBase)
	default:
		fmt.Printf("⚠️ 不支持的协议: %s\n", node.Protocol)
	}

	return result
}

// testV2RayNode 测试V2Ray节点
func (m *MVPTester) testV2RayNode(node *types.Node, result types.ValidNode, portBase int) types.ValidNode {
	fmt.Printf("  🔧 启动V2Ray代理测试...\n")

	proxyManager := proxy.NewProxyManager()
	defer func() {
		fmt.Printf("  🛑 清理V2Ray代理资源...\n")
		proxyManager.StopProxy()
	}()

	httpPort := portBase + 1
	socksPort := portBase + 2

	// 为每个测试线程创建唯一的配置文件名
	proxyManager.ConfigPath = fmt.Sprintf("temp_v2ray_config_%d_%d.json", portBase, time.Now().UnixNano())

	// 手动设置端口
	proxyManager.HTTPPort = httpPort
	proxyManager.SOCKSPort = socksPort

	fmt.Printf("  🔧 配置代理端口: HTTP=%d, SOCKS=%d\n", httpPort, socksPort)

	err := proxyManager.StartProxy(node)
	if err != nil {
		fmt.Printf("  ❌ V2Ray代理启动失败: %v\n", err)
		return result
	}

	// 等待代理启动 - Windows需要更长时间
	waitTime := 5 * time.Second
	if runtime.GOOS == "windows" {
		waitTime = 8 * time.Second
	}
	fmt.Printf("  ⏳ 等待代理启动 (%.0fs)...\n", waitTime.Seconds())
	time.Sleep(waitTime)

	// 验证代理是否真正启动
	if !m.verifyProxyStarted(httpPort) {
		fmt.Printf("  ❌ V2Ray代理启动验证失败\n")
		return result
	}

	// 测试连接性能
	proxyTestURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	fmt.Printf("  🧪 测试V2Ray代理URL: %s\n", proxyTestURL)

	latency, speed, err := m.testProxyPerformance(proxyTestURL)
	if err != nil {
		fmt.Printf("  ❌ V2Ray代理性能测试失败: %v\n", err)
		return result
	}

	// 计算综合分数 (速度权重70%，延迟权重30%)
	score := speed*0.7 + (1000.0/float64(latency))*0.3

	result.Node = node
	result.Latency = latency
	result.Speed = speed
	result.Score = score
	result.SuccessCount = 1

	fmt.Printf("  ✅ V2Ray节点测试成功\n")
	return result
}

// testHysteria2Node 测试Hysteria2节点
func (m *MVPTester) testHysteria2Node(node *types.Node, result types.ValidNode, portBase int) types.ValidNode {
	fmt.Printf("  🔧 启动Hysteria2代理测试...\n")

	hysteria2Manager := proxy.NewHysteria2ProxyManager()
	defer func() {
		fmt.Printf("  🛑 清理Hysteria2代理资源...\n")
		hysteria2Manager.StopHysteria2Proxy()
	}()

	httpPort := portBase + 1
	socksPort := portBase + 2

	// 为每个测试线程创建唯一的配置文件名
	hysteria2Manager.SetConfigPath(fmt.Sprintf("./hysteria2/config_%d_%d.yaml", portBase, time.Now().UnixNano()))

	// 手动设置端口
	hysteria2Manager.HTTPPort = httpPort
	hysteria2Manager.SOCKSPort = socksPort

	fmt.Printf("  🔧 配置代理端口: HTTP=%d, SOCKS=%d\n", httpPort, socksPort)

	err := hysteria2Manager.StartHysteria2Proxy(node)
	if err != nil {
		fmt.Printf("  ❌ Hysteria2代理启动失败: %v\n", err)
		return result
	}

	// 等待代理启动 - Windows需要更长时间
	waitTime := 5 * time.Second
	if runtime.GOOS == "windows" {
		waitTime = 10 * time.Second // Hysteria2在Windows下需要更长启动时间
	}
	fmt.Printf("  ⏳ 等待代理启动 (%.0fs)...\n", waitTime.Seconds())
	time.Sleep(waitTime)

	// 验证代理是否真正启动
	if !m.verifyProxyStarted(httpPort) {
		fmt.Printf("  ❌ Hysteria2代理启动验证失败\n")
		return result
	}

	// 测试连接性能
	proxyTestURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	fmt.Printf("  🧪 测试Hysteria2代理URL: %s\n", proxyTestURL)

	latency, speed, err := m.testProxyPerformance(proxyTestURL)
	if err != nil {
		fmt.Printf("  ❌ Hysteria2代理性能测试失败: %v\n", err)
		return result
	}

	// 计算综合分数
	score := speed*0.7 + (1000.0/float64(latency))*0.3

	result.Node = node
	result.Latency = latency
	result.Speed = speed
	result.Score = score
	result.SuccessCount = 1

	fmt.Printf("  ✅ Hysteria2节点测试成功\n")
	return result
}

// testProxyPerformance 测试代理性能
func (m *MVPTester) testProxyPerformance(proxyURL string) (int64, float64, error) {
	// 添加panic恢复机制
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ testProxyPerformance发生panic: %v\n", r)
		}
	}()

	// 检查输入参数
	if proxyURL == "" {
		return 0, 0, fmt.Errorf("代理URL为空")
	}

	// 创建代理客户端
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		return 0, 0, fmt.Errorf("解析代理URL失败: %v", err)
	}

	if proxyURLParsed == nil {
		return 0, 0, fmt.Errorf("解析后的代理URL为空")
	}

	// 创建代理函数
	proxyFunc := http.ProxyURL(proxyURLParsed)
	if proxyFunc == nil {
		return 0, 0, fmt.Errorf("创建代理函数失败")
	}

	// 使用配置中的超时时间，而不是硬编码
	var dialTimeout, clientTimeout time.Duration

	// 基于配置的超时时间计算各个阶段的超时
	configTimeout := m.testTimeout
	if configTimeout <= 0 {
		configTimeout = 30 * time.Second // 默认值
	}

	if runtime.GOOS == "windows" {
		// Windows环境使用配置的超时时间，但有最小值保证
		dialTimeout = configTimeout / 4
		if dialTimeout < 5*time.Second {
			dialTimeout = 5 * time.Second
		}
		clientTimeout = configTimeout
		if clientTimeout < 10*time.Second {
			clientTimeout = 10 * time.Second
		}
	} else {
		dialTimeout = configTimeout / 3
		clientTimeout = configTimeout
	}

	fmt.Printf("  ⏱️ 使用超时配置: 连接超时=%.0fs, 总超时=%.0fs\n",
		dialTimeout.Seconds(), clientTimeout.Seconds())

	// 创建更健壮的Transport配置
	transport := &http.Transport{
		Proxy: proxyFunc,
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 10 * time.Second, // 缩短Keep-Alive
		}).DialContext,
		ForceAttemptHTTP2:     false,           // 禁用HTTP/2，避免兼容性问题
		MaxIdleConns:          2,               // 进一步减少连接数
		IdleConnTimeout:       5 * time.Second, // 大幅缩短空闲超时
		TLSHandshakeTimeout:   8 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     true, // Windows下禁用Keep-Alive避免连接复用问题
		DisableCompression:    false,
		ResponseHeaderTimeout: 10 * time.Second, // 缩短响应头超时
	}

	// 检查transport是否创建成功
	if transport == nil {
		return 0, 0, fmt.Errorf("创建传输层失败")
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   clientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("禁止重定向") // 禁止重定向，简化测试
		},
	}

	// 检查client是否创建成功
	if client == nil {
		return 0, 0, fmt.Errorf("创建HTTP客户端失败")
	}

	// 使用配置中的测试URL，如果没有配置则使用默认值
	var testURLs []string
	if m.testURL != "" {
		// 用户配置了测试URL，优先使用
		testURLs = []string{m.testURL}
		fmt.Printf("  🎯 使用配置的测试URL: %s\n", m.testURL)
	} else if runtime.GOOS == "windows" {
		// Windows环境使用更简单、更快的测试URL
		testURLs = []string{
			"http://httpbin.org/get?test=1",               // 简单GET请求
			"http://www.baidu.com/robots.txt",             // 小文件，国内快速
			"http://captive.apple.com/hotspot-detect.txt", // 苹果连通性检测
		}
		fmt.Printf("  🪟 Windows环境：使用优化的测试URL列表\n")
	} else {
		testURLs = []string{
			"http://httpbin.org/ip",
			"http://www.google.com",
		}
		fmt.Printf("  🌐 Unix环境：使用标准测试URL列表\n")
	}

	var lastErr error
	for i, testURL := range testURLs {
		fmt.Printf("🔍 尝试测试URL [%d/%d]: %s\n", i+1, len(testURLs), testURL)

		// 为每个URL创建带超时的context - 使用更短的超时
		shortTimeout := clientTimeout / 2 // 每个URL只用一半时间
		ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)
		defer cancel()

		// Windows下只尝试一次，避免浪费时间
		maxRetries := 1
		if runtime.GOOS != "windows" {
			maxRetries = 2
		}

		var resp *http.Response
		var err error
		var start time.Time

		for attempt := 1; attempt <= maxRetries; attempt++ {
			fmt.Printf("  🔄 尝试 %d/%d (超时%.0fs)...\n", attempt, maxRetries, shortTimeout.Seconds())

			// 测试延迟
			start = time.Now()

			// 创建带context的请求
			req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
			if err != nil {
				lastErr = fmt.Errorf("创建请求失败: %v", err)
				break
			}

			if req == nil {
				lastErr = fmt.Errorf("创建的请求为空")
				break
			}

			// 设置最简化的请求头
			req.Header.Set("User-Agent", "test/1.0")
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Connection", "close")

			// 检查客户端是否为空
			if client == nil {
				lastErr = fmt.Errorf("HTTP客户端为空")
				break
			}

			// 设置请求开始时间用于超时检测
			requestStart := time.Now()

			resp, err = client.Do(req)

			// 检查是否超时
			if time.Since(requestStart) > shortTimeout {
				if resp != nil && resp.Body != nil {
					resp.Body.Close()
				}
				lastErr = fmt.Errorf("请求超时 (%.1fs)", time.Since(requestStart).Seconds())
				fmt.Printf("  ⏰ 请求超时，跳过\n")
				break
			}

			if err == nil {
				break // 成功，跳出重试循环
			}

			lastErr = fmt.Errorf("请求失败: %v", err)
			fmt.Printf("  ❌ %v\n", lastErr)

			// 如果不是最后一次尝试，短暂等待再重试
			if attempt < maxRetries {
				time.Sleep(500 * time.Millisecond) // 缩短重试间隔
			}
		}

		if err != nil {
			fmt.Printf("  ❌ URL %s 失败，尝试下一个\n", testURL)
			continue // 这个URL失败，尝试下一个
		}

		// 检查响应是否为空
		if resp == nil {
			lastErr = fmt.Errorf("响应对象为空")
			fmt.Printf("  ❌ %v\n", lastErr)
			continue
		}

		latency := time.Since(start).Milliseconds()

		// 接受更多状态码，提高成功率
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			if resp.Body != nil {
				resp.Body.Close()
			}
			lastErr = fmt.Errorf("状态码: %d", resp.StatusCode)
			fmt.Printf("  ❌ %v，尝试下一个URL\n", lastErr)
			continue
		}

		// 简化速度测试 - 限制读取大小和时间
		speedStart := time.Now()

		// 检查响应体是否为空
		if resp.Body == nil {
			lastErr = fmt.Errorf("响应体为空")
			fmt.Printf("  ❌ %v\n", lastErr)
			continue
		}

		// 限制读取大小，避免下载过大内容
		maxReadSize := int64(64 * 1024) // 最多读取64KB，减少读取量
		limitedReader := io.LimitReader(resp.Body, maxReadSize)

		// 设置更短的读取超时
		readTimeout := 5 * time.Second
		if runtime.GOOS == "windows" {
			readTimeout = 8 * time.Second
		}

		readCtx, readCancel := context.WithTimeout(context.Background(), readTimeout)
		defer readCancel()

		// 在goroutine中读取，避免阻塞
		type readResult struct {
			data []byte
			err  error
		}

		readChan := make(chan readResult, 1)
		go func() {
			data, err := io.ReadAll(limitedReader)
			readChan <- readResult{data: data, err: err}
		}()

		var body []byte
		select {
		case result := <-readChan:
			body = result.data
			err = result.err
		case <-readCtx.Done():
			err = fmt.Errorf("读取响应超时")
		}

		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %v", err)
			fmt.Printf("  ❌ %v\n", lastErr)
			continue
		}

		downloadTime := time.Since(speedStart).Seconds()
		if downloadTime == 0 {
			downloadTime = 0.001 // 避免除零
		}

		// 计算速度 (bytes/s -> Mbps)
		speed := float64(len(body)) / downloadTime / 1024 / 1024 * 8

		fmt.Printf("  ✅ 代理测试成功 - URL: %s, 延迟: %dms, 大小: %d bytes, 速度: %.2f Mbps\n",
			testURL, latency, len(body), speed)

		return latency, speed, nil
	}

	return 0, 0, fmt.Errorf("所有测试URL都失败，最后错误: %v", lastErr)
}

// showTestSummary 显示测试摘要
func (m *MVPTester) showTestSummary(validNodes []types.ValidNode) {
	fmt.Printf("\n📊 测试摘要:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if len(validNodes) == 0 {
		fmt.Printf("❌ 没有有效节点\n")
		return
	}

	// 显示前5个最快的节点
	displayCount := 5
	if len(validNodes) < displayCount {
		displayCount = len(validNodes)
	}

	for i := 0; i < displayCount; i++ {
		node := validNodes[i]
		fmt.Printf("🏆 #%d %s (分数: %.2f, 延迟: %dms, 速度: %.2fMbps)\n",
			i+1, node.Node.Name, node.Score, node.Latency, node.Speed)
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// loadBestNode 加载最佳节点
func (m *MVPTester) loadBestNode() {
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		fmt.Printf("📄 没有找到历史最佳节点文件\n")
		return
	}

	var state MVPState
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("⚠️ 解析历史最佳节点失败: %v\n", err)
		return
	}

	m.mutex.Lock()
	m.bestNode = state.BestNode
	m.mutex.Unlock()

	if m.bestNode != nil {
		fmt.Printf("📚 加载历史最佳节点: %s (分数: %.2f)\n",
			m.bestNode.Node.Name, m.bestNode.Score)
	}
}

// saveBestNode 保存最佳节点
func (m *MVPTester) saveBestNode() error {
	m.mutex.RLock()
	state := MVPState{
		BestNode:   m.bestNode,
		LastUpdate: time.Now(),
	}
	m.mutex.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.stateFile, data, 0644)
}

// setupSignalHandler 设置信号处理
func (m *MVPTester) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("\n🛑 接收到退出信号，正在清理资源...\n")
		m.Stop()
		os.Exit(0)
	}()
}

// checkDependencies 检查依赖
func (m *MVPTester) checkDependencies() error {
	fmt.Printf("🔧 检查依赖...\n")

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

	fmt.Printf("✅ 所有依赖检查完成\n")
	return nil
}

// GetBestNode 获取当前最佳节点
func (m *MVPTester) GetBestNode() *types.ValidNode {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.bestNode
}

// RunMVPTester 运行MVP测试器
func RunMVPTester(subscriptionURL string) error {
	tester := NewMVPTester(subscriptionURL)
	return tester.Start()
}

// verifyProxyStarted 验证代理是否成功启动
func (m *MVPTester) verifyProxyStarted(port int) bool {
	// 尝试连接到代理端口
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 3*time.Second)
	if err != nil {
		fmt.Printf("    ❌ 无法连接到代理端口 %d: %v\n", port, err)
		return false
	}
	conn.Close()
	fmt.Printf("    ✅ 代理端口 %d 连接正常\n", port)
	return true
}
