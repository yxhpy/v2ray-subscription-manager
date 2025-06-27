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

	// 限制并发数
	concurrency := 3 // 减少并发数以提高成功率
	semaphore := make(chan struct{}, concurrency)

	for i, node := range nodes {
		wg.Add(1)
		go func(node *types.Node, index int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("🧪 测试节点 [%d/%d]: %s (%s)\n",
				index+1, len(nodes), node.Name, node.Protocol)

			validNode := m.testSingleNode(node, 8000+index*10)
			if validNode.Node != nil {
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
				fmt.Printf("❌ 节点 %s 测试失败\n", node.Name)
			}
		}(node, i)
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
	proxyManager := proxy.NewProxyManager()
	defer proxyManager.StopProxy()

	httpPort := portBase + 1
	socksPort := portBase + 2

	// 为每个测试线程创建唯一的配置文件名
	proxyManager.ConfigPath = fmt.Sprintf("temp_v2ray_config_%d_%d.json", portBase, time.Now().UnixNano())

	// 手动设置端口
	proxyManager.HTTPPort = httpPort
	proxyManager.SOCKSPort = socksPort

	err := proxyManager.StartProxy(node)
	if err != nil {
		return result
	}

	// 等待代理启动
	time.Sleep(5 * time.Second)

	// 测试连接性能
	proxyTestURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	fmt.Printf("🧪 测试V2Ray代理URL: %s\n", proxyTestURL)

	latency, speed, err := m.testProxyPerformance(proxyTestURL)
	if err != nil {
		fmt.Printf("❌ V2Ray代理性能测试失败: %v\n", err)
		return result
	}

	// 计算综合分数 (速度权重70%，延迟权重30%)
	score := speed*0.7 + (1000.0/float64(latency))*0.3

	result.Node = node
	result.Latency = latency
	result.Speed = speed
	result.Score = score
	result.SuccessCount = 1

	return result
}

// testHysteria2Node 测试Hysteria2节点
func (m *MVPTester) testHysteria2Node(node *types.Node, result types.ValidNode, portBase int) types.ValidNode {
	hysteria2Manager := proxy.NewHysteria2ProxyManager()
	defer hysteria2Manager.StopHysteria2Proxy()

	httpPort := portBase + 1
	socksPort := portBase + 2

	// 为每个测试线程创建唯一的配置文件名
	hysteria2Manager.SetConfigPath(fmt.Sprintf("./hysteria2/config_%d_%d.yaml", portBase, time.Now().UnixNano()))

	// 手动设置端口
	hysteria2Manager.HTTPPort = httpPort
	hysteria2Manager.SOCKSPort = socksPort

	err := hysteria2Manager.StartHysteria2Proxy(node)
	if err != nil {
		return result
	}

	// 等待代理启动 - Windows需要更长时间
	waitTime := 5 * time.Second
	if runtime.GOOS == "windows" {
		waitTime = 8 * time.Second
	}
	time.Sleep(waitTime)

	// 测试连接性能
	proxyTestURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	fmt.Printf("🧪 测试Hysteria2代理URL: %s\n", proxyTestURL)

	latency, speed, err := m.testProxyPerformance(proxyTestURL)
	if err != nil {
		fmt.Printf("❌ Hysteria2代理性能测试失败: %v\n", err)
		return result
	}

	// 计算综合分数
	score := speed*0.7 + (1000.0/float64(latency))*0.3

	result.Node = node
	result.Latency = latency
	result.Speed = speed
	result.Score = score
	result.SuccessCount = 1

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

	// 创建更健壮的Transport配置
	transport := &http.Transport{
		Proxy: proxyFunc,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false, // 禁用HTTP/2，避免兼容性问题
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}

	// 检查transport是否创建成功
	if transport == nil {
		return 0, 0, fmt.Errorf("创建传输层失败")
	}

	// Windows下使用更长的超时时间
	timeout := 20 * time.Second
	if runtime.GOOS == "windows" {
		timeout = 45 * time.Second
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}

	// 检查client是否创建成功
	if client == nil {
		return 0, 0, fmt.Errorf("创建HTTP客户端失败")
	}

	// 根据系统环境选择测试URL
	var testURLs []string
	if runtime.GOOS == "windows" {
		// Windows环境优先使用国内和稳定的URL
		testURLs = []string{
			"http://www.baidu.com",
			"http://httpbin.org/ip",
			"http://www.bing.com",
			"http://www.github.com",
			"http://www.google.com", // 放到最后尝试
		}
	} else {
		// Unix环境使用原有策略
		testURLs = []string{
			"http://httpbin.org/ip",
			"http://www.google.com",
			"http://www.baidu.com",
			"http://www.github.com",
		}
	}

	var lastErr error
	for _, testURL := range testURLs {
		// 对每个URL进行重试
		maxRetries := 2
		if runtime.GOOS == "windows" {
			maxRetries = 3 // Windows下增加重试次数
		}

		var resp *http.Response
		var err error
		var start time.Time

		for attempt := 1; attempt <= maxRetries; attempt++ {
			// 测试延迟
			start = time.Now()

			// 创建请求
			req, err := http.NewRequest("GET", testURL, nil)
			if err != nil {
				lastErr = fmt.Errorf("创建请求失败: %v", err)
				break
			}

			if req == nil {
				lastErr = fmt.Errorf("创建的请求为空")
				break
			}

			// 设置更兼容的请求头
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
			req.Header.Set("Accept-Encoding", "gzip, deflate")
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Cache-Control", "no-cache")

			// 检查客户端是否为空
			if client == nil {
				lastErr = fmt.Errorf("HTTP客户端为空")
				break
			}

			resp, err = client.Do(req)
			if err == nil {
				break // 成功，跳出重试循环
			}

			lastErr = fmt.Errorf("请求 %s 失败 (尝试 %d/%d): %v", testURL, attempt, maxRetries, err)

			// 如果不是最后一次尝试，等待一段时间再重试
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}

		if err != nil {
			continue // 这个URL失败，尝试下一个
		}

		// 检查响应是否为空
		if resp == nil {
			lastErr = fmt.Errorf("响应对象为空")
			continue
		}

		latency := time.Since(start).Milliseconds()

		if resp.StatusCode != http.StatusOK {
			if resp.Body != nil {
				resp.Body.Close()
			}
			lastErr = fmt.Errorf("%s 返回状态码: %d", testURL, resp.StatusCode)
			continue
		}

		// 测试速度 - 下载内容
		speedStart := time.Now()

		// 检查响应体是否为空
		if resp.Body == nil {
			lastErr = fmt.Errorf("响应体为空")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("读取 %s 响应失败: %v", testURL, err)
			continue
		}

		downloadTime := time.Since(speedStart).Seconds()
		if downloadTime == 0 {
			downloadTime = 0.001 // 避免除零
		}

		// 计算速度 (bytes/s -> Mbps)
		speed := float64(len(body)) / downloadTime / 1024 / 1024 * 8

		fmt.Printf("🌐 代理测试成功 - URL: %s, 延迟: %dms, 大小: %d bytes, 速度: %.2f Mbps\n",
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
