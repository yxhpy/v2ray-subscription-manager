package workflow

import (
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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/core/downloader"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/parser"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// SpeedTestResult 测速结果
type SpeedTestResult struct {
	Node     *types.Node `json:"node"`
	Success  bool        `json:"success"`
	Latency  int64       `json:"latency_ms"` // 延迟毫秒
	Error    string      `json:"error,omitempty"`
	TestTime time.Time   `json:"test_time"`
	Speed    float64     `json:"speed_mbps"` // 速度 Mbps
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	SubscriptionURL string `json:"subscription_url"`
	MaxConcurrency  int    `json:"max_concurrency"`
	TestTimeout     int    `json:"test_timeout_seconds"`
	OutputFile      string `json:"output_file"`
	TestURL         string `json:"test_url"`
	MaxNodes        int    `json:"max_nodes"` // 最大测试节点数
}

// SpeedTestWorkflow 测速工作流
type SpeedTestWorkflow struct {
	config         WorkflowConfig
	results        []SpeedTestResult
	mutex          sync.Mutex
	activeManagers []ProxyManagerInterface // 跟踪活跃的代理管理器
	managerMutex   sync.Mutex
}

// ProxyManagerInterface 代理管理器接口
type ProxyManagerInterface interface {
	Stop() error
}

// ProxyManagerWrapper V2Ray代理管理器包装器
type ProxyManagerWrapper struct {
	*proxy.ProxyManager
}

func (p *ProxyManagerWrapper) Stop() error {
	return p.StopProxy()
}

// Hysteria2ProxyManagerWrapper Hysteria2代理管理器包装器
type Hysteria2ProxyManagerWrapper struct {
	*proxy.Hysteria2ProxyManager
}

func (h *Hysteria2ProxyManagerWrapper) Stop() error {
	return h.StopHysteria2Proxy()
}

// NewSpeedTestWorkflow 创建新的测速工作流
func NewSpeedTestWorkflow(subscriptionURL string) *SpeedTestWorkflow {
	return &SpeedTestWorkflow{
		config: WorkflowConfig{
			SubscriptionURL: subscriptionURL,
			MaxConcurrency:  10, // 降低到10个并发，避免资源耗尽
			TestTimeout:     30, // 增加到30秒超时，适应Windows环境
			OutputFile:      "speed_test_results.txt",
			TestURL:         "http://www.baidu.com", // 默认使用百度
			MaxNodes:        0,                      // 0表示不限制
		},
		results:        make([]SpeedTestResult, 0),
		activeManagers: make([]ProxyManagerInterface, 0),
	}
}

// SetConcurrency 设置并发数
func (w *SpeedTestWorkflow) SetConcurrency(concurrency int) {
	w.config.MaxConcurrency = concurrency
}

// SetTimeout 设置超时时间
func (w *SpeedTestWorkflow) SetTimeout(timeout int) {
	w.config.TestTimeout = timeout
}

// SetOutputFile 设置输出文件
func (w *SpeedTestWorkflow) SetOutputFile(filename string) {
	w.config.OutputFile = filename
}

// SetTestURL 设置测试URL
func (w *SpeedTestWorkflow) SetTestURL(url string) {
	w.config.TestURL = url
}

// SetMaxNodes 设置最大测试节点数
func (w *SpeedTestWorkflow) SetMaxNodes(maxNodes int) {
	w.config.MaxNodes = maxNodes
}

// Run 运行工作流
func (w *SpeedTestWorkflow) Run() error {
	fmt.Printf("🚀 开始执行测速工作流...\n")
	fmt.Printf("📡 订阅链接: %s\n", w.config.SubscriptionURL)
	fmt.Printf("⚡ 并发数: %d\n", w.config.MaxConcurrency)
	fmt.Printf("⏱️  超时时间: %d秒\n", w.config.TestTimeout)
	fmt.Printf("🎯 测试目标: %s\n", w.config.TestURL)
	fmt.Printf("📄 输出文件: %s\n", w.config.OutputFile)

	// 设置信号处理，确保程序退出时清理资源
	w.setupSignalHandler()

	// 步骤0: 检查和安装依赖
	fmt.Printf("\n🔧 检查和安装必要依赖...\n")
	err := w.checkAndInstallDependencies()
	if err != nil {
		return fmt.Errorf("依赖检查失败: %v", err)
	}
	fmt.Printf("✅ 所有依赖已就绪\n")

	// 步骤1: 解析订阅链接
	fmt.Printf("\n📥 正在解析订阅链接...\n")
	nodes, err := w.parseSubscription()
	if err != nil {
		return fmt.Errorf("解析订阅失败: %v", err)
	}
	fmt.Printf("✅ 成功解析 %d 个节点\n", len(nodes))

	// 步骤2: 多线程测试所有节点
	fmt.Printf("\n🧪 开始多线程测试节点...\n")
	fmt.Printf("💪 使用 %d 个线程并发测试，榨干CPU性能！\n", w.config.MaxConcurrency)
	err = w.testAllNodes(nodes)
	if err != nil {
		return fmt.Errorf("测试节点失败: %v", err)
	}

	// 步骤3: 按速度排序
	fmt.Printf("\n📊 按速度排序结果...\n")
	w.sortResultsBySpeed()

	// 步骤4: 写入文件
	fmt.Printf("\n💾 保存结果到文件...\n")
	err = w.saveResults()
	if err != nil {
		return fmt.Errorf("保存结果失败: %v", err)
	}

	// 显示摘要
	w.showSummary()

	// 最终清理
	w.cleanupAllResources()

	// 额外的深度清理
	w.deepCleanup()

	fmt.Printf("\n🎉 工作流执行完成！\n")
	return nil
}

// setupSignalHandler 设置信号处理器
func (w *SpeedTestWorkflow) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("\n🛑 接收到退出信号，正在清理资源...\n")
		w.cleanupAllResources()
		os.Exit(1)
	}()
}

// addActiveManager 添加活跃的代理管理器
func (w *SpeedTestWorkflow) addActiveManager(manager ProxyManagerInterface) {
	w.managerMutex.Lock()
	defer w.managerMutex.Unlock()
	w.activeManagers = append(w.activeManagers, manager)
}

// removeActiveManager 从活跃管理器列表中移除
func (w *SpeedTestWorkflow) removeActiveManager(manager ProxyManagerInterface) {
	w.managerMutex.Lock()
	defer w.managerMutex.Unlock()

	for i, m := range w.activeManagers {
		if m == manager {
			// 从切片中移除元素
			w.activeManagers = append(w.activeManagers[:i], w.activeManagers[i+1:]...)
			break
		}
	}
}

// cleanupAllResources 清理所有资源
func (w *SpeedTestWorkflow) cleanupAllResources() {
	fmt.Printf("🧹 清理所有活跃的代理进程...\n")
	w.managerMutex.Lock()
	defer w.managerMutex.Unlock()

	for _, manager := range w.activeManagers {
		manager.Stop()
	}
	w.activeManagers = nil

	// 强制杀掉所有可能的残留进程
	exec.Command("pkill", "-f", "v2ray").Run()
	exec.Command("pkill", "-f", "hysteria").Run()
	fmt.Printf("✅ 资源清理完成\n")
}

// deepCleanup 深度清理资源
func (w *SpeedTestWorkflow) deepCleanup() {
	fmt.Printf("🧹 执行深度资源清理...\n")

	// 清理所有可能的临时配置文件
	if runtime.GOOS != "windows" {
		// Unix/Linux/macOS环境下的清理
		exec.Command("find", ".", "-name", "temp_config_*.json", "-delete").Run()
		exec.Command("find", ".", "-name", "config_*.yaml", "-delete").Run()
		exec.Command("rm", "-f", "hysteria2/config.yaml.tmp*").Run()
		exec.Command("rm", "-f", "hysteria2/config_*.yaml").Run()

		// 强制清理所有可能占用的端口（轻量级检查）
		for port := 10000; port < 20000; port += 100 {
			// 只检查主要端口，不执行kill操作避免影响其他进程
			exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port)).Run()
		}

		// 最后一次强制清理进程
		exec.Command("pkill", "-f", "v2ray").Run()
		exec.Command("pkill", "-f", "hysteria").Run()

		fmt.Printf("🧹 Unix环境清理完成\n")
	} else {
		// Windows环境下的清理
		w.cleanupTempFilesWindows()
	}

	// 跨平台通用清理
	w.cleanupAdditionalTempFiles()

	// 等待一下让进程完全退出
	time.Sleep(2 * time.Second)

	fmt.Printf("✅ 深度清理完成\n")
}

// cleanupTempFilesWindows Windows环境下的临时文件清理
func (w *SpeedTestWorkflow) cleanupTempFilesWindows() {
	fmt.Printf("🧹 Windows环境临时文件清理...\n")

	// 清理V2Ray临时配置文件
	files, err := filepath.Glob("temp_config_*.json")
	if err == nil {
		for _, file := range files {
			if err := os.Remove(file); err == nil {
				fmt.Printf("🧹 已清理V2Ray配置: %s\n", file)
			}
		}
	}

	// 调用专门的Hysteria2清理方法
	w.cleanupWindowsHysteria2Files()

	// 额外清理可能遗留的文件
	w.cleanupAdditionalTempFiles()
}

// cleanupAdditionalTempFiles 清理额外的临时文件
func (w *SpeedTestWorkflow) cleanupAdditionalTempFiles() {
	// 清理可能的其他临时文件模式
	patterns := []string{
		"*.tmp",
		"*.temp",
		"config_*.json",
		"temp_*.yaml",
		"test_proxy_*.json", // 添加test_proxy_开头的JSON文件
		"test_proxy_*.yaml", // 添加test_proxy_开头的YAML文件
	}

	for _, pattern := range patterns {
		if files, err := filepath.Glob(pattern); err == nil {
			for _, file := range files {
				// 只删除明显是临时文件的
				if strings.Contains(file, "temp") || strings.Contains(file, "tmp") || strings.Contains(file, "test_proxy") {
					if err := os.Remove(file); err == nil {
						fmt.Printf("🧹 已清理临时文件: %s\n", file)
					}
				}
			}
		}
	}
}

// cleanupHysteria2TempFiles 清理Hysteria2临时文件
func (w *SpeedTestWorkflow) cleanupHysteria2TempFiles(manager *proxy.Hysteria2ProxyManager) {
	if manager == nil {
		return
	}

	// 由于downloader字段是私有的，我们直接清理可能的临时配置文件

	// 清理可能的临时配置文件（使用多种模式匹配）
	patterns := []string{
		"./hysteria2/config_*.yaml",    // 新的命名模式
		"./hysteria2/config.yaml.tmp*", // 可能的临时文件
		"hysteria2/config_*.yaml",      // 无./前缀的模式
		"hysteria2/config.yaml.tmp*",   // 无./前缀的临时文件
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err == nil {
			for _, file := range files {
				if err := os.Remove(file); err == nil {
					fmt.Printf("🧹 已清理临时文件: %s\n", file)
				}
			}
		}
	}

	// Windows特殊处理：强制清理可能被锁定的文件
	if runtime.GOOS == "windows" {
		w.cleanupWindowsHysteria2Files()
	}
}

// cleanupWindowsHysteria2Files Windows下的特殊清理方法
func (w *SpeedTestWorkflow) cleanupWindowsHysteria2Files() {
	// 等待一小段时间，让文件句柄释放
	time.Sleep(100 * time.Millisecond)

	// 尝试清理hysteria2目录下的所有yaml文件
	hysteria2Dir := "./hysteria2"
	if _, err := os.Stat(hysteria2Dir); err == nil {
		files, err := filepath.Glob(filepath.Join(hysteria2Dir, "*.yaml"))
		if err == nil {
			for _, file := range files {
				// 检查文件是否包含临时标识
				if strings.Contains(file, "config_") || strings.Contains(file, ".tmp") {
					// 多次尝试删除，因为Windows可能有文件锁
					for i := 0; i < 3; i++ {
						if err := os.Remove(file); err == nil {
							fmt.Printf("🧹 Windows清理成功: %s\n", file)
							break
						} else if i == 2 {
							fmt.Printf("⚠️  Windows清理失败 %s: %v\n", file, err)
						} else {
							time.Sleep(50 * time.Millisecond)
						}
					}
				}
			}
		}
	}
}

// checkAndInstallDependencies 检查和安装必要依赖
func (w *SpeedTestWorkflow) checkAndInstallDependencies() error {
	fmt.Printf("🔍 检查V2Ray核心...\n")
	v2rayDownloader := downloader.NewV2RayDownloader()
	if !v2rayDownloader.CheckV2rayInstalled() {
		fmt.Printf("📥 V2Ray未安装，正在下载...\n")
		if err := downloader.AutoDownloadV2Ray(); err != nil {
			return fmt.Errorf("V2Ray下载失败: %v", err)
		}
		fmt.Printf("✅ V2Ray安装完成\n")
	} else {
		fmt.Printf("✅ V2Ray已安装\n")
	}

	fmt.Printf("🔍 检查Hysteria2客户端...\n")
	hysteria2Downloader := downloader.NewHysteria2Downloader()
	if !hysteria2Downloader.CheckHysteria2Installed() {
		fmt.Printf("📥 Hysteria2未安装，正在下载...\n")
		if err := downloader.AutoDownloadHysteria2(); err != nil {
			return fmt.Errorf("Hysteria2下载失败: %v", err)
		}
		fmt.Printf("✅ Hysteria2安装完成\n")
	} else {
		fmt.Printf("✅ Hysteria2已安装\n")
	}

	return nil
}

// parseSubscription 解析订阅链接
func (w *SpeedTestWorkflow) parseSubscription() ([]*types.Node, error) {
	// 获取订阅内容
	content, err := parser.FetchSubscription(w.config.SubscriptionURL)
	if err != nil {
		return nil, err
	}

	// Base64解码
	decodedContent, err := parser.DecodeBase64(content)
	if err != nil {
		return nil, err
	}

	// 解析链接
	nodes, err := parser.ParseLinks(decodedContent)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到有效节点")
	}

	// 如果设置了最大节点数限制，只取前N个节点
	if w.config.MaxNodes > 0 && len(nodes) > w.config.MaxNodes {
		nodes = nodes[:w.config.MaxNodes]
		fmt.Printf("⚠️  限制测试节点数为 %d 个\n", w.config.MaxNodes)
	}

	return nodes, nil
}

// testAllNodes 多线程测试所有节点
func (w *SpeedTestWorkflow) testAllNodes(nodes []*types.Node) error {
	// 创建工作队列
	nodeQueue := make(chan *types.Node, len(nodes))
	resultQueue := make(chan SpeedTestResult, len(nodes))

	// 填充工作队列
	for _, node := range nodes {
		nodeQueue <- node
	}
	close(nodeQueue)

	// 创建工作协程，为每个协程分配不同的端口范围
	var wg sync.WaitGroup
	for i := 0; i < w.config.MaxConcurrency; i++ {
		wg.Add(1)
		// 为每个worker分配不同的端口基数，避免端口冲突
		portBase := 10000 + i*100 // worker 0: 10000-10099, worker 1: 10100-10199, 等等
		go w.worker(nodeQueue, resultQueue, &wg, portBase)
	}

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(resultQueue)
	}()

	// 收集结果
	totalNodes := len(nodes)
	completed := 0
	for result := range resultQueue {
		w.mutex.Lock()
		w.results = append(w.results, result)
		completed++
		w.mutex.Unlock()

		// 显示进度
		fmt.Printf("\r🔄 测试进度: %d/%d (%.1f%%) - 最新: %s",
			completed, totalNodes, float64(completed)/float64(totalNodes)*100, result.Node.Name)
	}

	fmt.Printf("\n✅ 测试完成，共测试 %d 个节点\n", len(w.results))
	return nil
}

// worker 工作协程
func (w *SpeedTestWorkflow) worker(nodeQueue <-chan *types.Node, resultQueue chan<- SpeedTestResult, wg *sync.WaitGroup, portBase int) {
	defer wg.Done()

	for node := range nodeQueue {
		result := w.testSingleNode(node, portBase)
		resultQueue <- result
	}
}

// testSingleNode 测试单个节点
func (w *SpeedTestWorkflow) testSingleNode(node *types.Node, portBase int) SpeedTestResult {
	result := SpeedTestResult{
		Node:     node,
		Success:  false,
		TestTime: time.Now(),
	}

	// 根据协议选择不同的代理方式
	if node.Protocol == "hysteria2" {
		return w.testHysteria2Node(node, result, portBase)
	} else {
		return w.testV2RayNode(node, result, portBase)
	}
}

// testV2RayNode 使用V2Ray测试节点
func (w *SpeedTestWorkflow) testV2RayNode(node *types.Node, result SpeedTestResult, portBase int) SpeedTestResult {
	// 创建临时V2Ray代理管理器
	tempManager := proxy.NewProxyManager()
	tempManager.ConfigPath = fmt.Sprintf("temp_config_%s_%d.json", node.Protocol, time.Now().UnixNano())

	// 设置专用端口，避免冲突
	tempManager.HTTPPort = portBase + 1  // HTTP代理端口
	tempManager.SOCKSPort = portBase + 2 // SOCKS代理端口

	// 添加到活跃管理器列表（使用包装器）
	wrapper := &ProxyManagerWrapper{tempManager}
	w.addActiveManager(wrapper)

	// 确保资源完全清理
	defer func() {
		// 停止代理
		tempManager.StopProxy()
		// 从活跃管理器列表中移除
		w.removeActiveManager(wrapper)
		// 清理临时配置文件
		os.Remove(tempManager.ConfigPath)
		// 强制清理可能的残留进程
		exec.Command("pkill", "-f", fmt.Sprintf(":%d", tempManager.HTTPPort)).Run()
		exec.Command("pkill", "-f", fmt.Sprintf(":%d", tempManager.SOCKSPort)).Run()
	}()

	// 启动V2Ray代理
	err := tempManager.StartProxy(node)
	if err != nil {
		result.Error = fmt.Sprintf("启动V2Ray代理失败: %v", err)
		return result
	}

	// Windows环境需要更长的启动时间
	waitTime := 2 * time.Second
	if runtime.GOOS == "windows" {
		waitTime = 5 * time.Second
	}
	time.Sleep(waitTime)

	// 测试连接和速度
	latency, speed, err := w.testProxySpeed(tempManager.HTTPPort)
	if err != nil {
		result.Error = fmt.Sprintf("测试失败: %v", err)
		return result
	}

	result.Success = true
	result.Latency = latency
	result.Speed = speed

	return result
}

// testHysteria2Node 使用Hysteria2客户端测试节点
func (w *SpeedTestWorkflow) testHysteria2Node(node *types.Node, result SpeedTestResult, portBase int) SpeedTestResult {
	// 创建临时Hysteria2代理管理器
	tempHysteria2Manager := proxy.NewHysteria2ProxyManager()

	// 设置专用端口，避免冲突
	tempHysteria2Manager.HTTPPort = portBase + 3  // HTTP代理端口
	tempHysteria2Manager.SOCKSPort = portBase + 4 // SOCKS代理端口

	// 添加到活跃管理器列表（使用包装器）
	wrapper := &Hysteria2ProxyManagerWrapper{tempHysteria2Manager}
	w.addActiveManager(wrapper)

	// 确保资源完全清理
	defer func() {
		// 停止Hysteria2代理
		tempHysteria2Manager.StopHysteria2Proxy()
		// 从活跃管理器列表中移除
		w.removeActiveManager(wrapper)
		// 强制清理可能的残留进程
		if runtime.GOOS != "windows" {
			exec.Command("pkill", "-f", fmt.Sprintf(":%d", tempHysteria2Manager.HTTPPort)).Run()
			exec.Command("pkill", "-f", fmt.Sprintf(":%d", tempHysteria2Manager.SOCKSPort)).Run()
		}
		// 清理临时配置文件
		w.cleanupHysteria2TempFiles(tempHysteria2Manager)
	}()

	// 启动Hysteria2代理
	err := tempHysteria2Manager.StartHysteria2Proxy(node)
	if err != nil {
		result.Error = fmt.Sprintf("启动Hysteria2代理失败: %v", err)
		return result
	}

	// Windows环境需要更长的启动时间
	waitTime := 2 * time.Second
	if runtime.GOOS == "windows" {
		waitTime = 5 * time.Second
	}
	time.Sleep(waitTime)

	// 测试连接和速度
	latency, speed, err := w.testProxySpeed(tempHysteria2Manager.HTTPPort)
	if err != nil {
		result.Error = fmt.Sprintf("测试失败: %v", err)
		return result
	}

	result.Success = true
	result.Latency = latency
	result.Speed = speed

	return result
}

// testProxySpeed 测试代理速度
func (w *SpeedTestWorkflow) testProxySpeed(proxyPort int) (int64, float64, error) {
	// 创建HTTP客户端 - 针对Windows环境优化
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", proxyPort)

	// 创建更健壮的Transport配置
	transport := &http.Transport{
		Proxy: http.ProxyURL(mustParseURL(proxyURL)),
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(w.config.TestTimeout) * time.Second,
			KeepAlive: 30 * time.Second, // 保持连接活跃
		}).DialContext,
		ForceAttemptHTTP2:     false, // 禁用HTTP/2，避免兼容性问题
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false, // 允许Keep-Alive
		DisableCompression:    false, // 允许压缩
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(w.config.TestTimeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 限制重定向次数，避免无限重定向
			if len(via) >= 3 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}

	// 测试延迟 - 增加重试机制
	var resp *http.Response
	var latency int64
	var err error

	// 重试最多3次
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		startTime := time.Now()

		// 简化逻辑：去掉重试中的context，使用client自带的超时
		req, err := http.NewRequest("GET", w.config.TestURL, nil)
		if err != nil {
			if attempt == maxRetries {
				return 0, 0, fmt.Errorf("创建请求失败: %v", err)
			}
			time.Sleep(time.Duration(attempt) * time.Second) // 递增等待时间
			continue
		}

		// 设置更兼容的User-Agent
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Connection", "keep-alive")

		resp, err = client.Do(req)

		if err != nil {
			if attempt == maxRetries {
				// 提供更详细的错误信息
				if strings.Contains(err.Error(), "unexpected EOF") {
					return 0, 0, fmt.Errorf("连接意外中断，可能是代理配置问题或网络不稳定")
				} else if strings.Contains(err.Error(), "timeout") {
					return 0, 0, fmt.Errorf("连接超时，请检查网络连接或增加超时时间")
				} else if strings.Contains(err.Error(), "connection refused") {
					return 0, 0, fmt.Errorf("连接被拒绝，代理服务可能未正常启动")
				} else if strings.Contains(err.Error(), "context canceled") {
					return 0, 0, fmt.Errorf("连接被取消，可能是网络超时")
				}
				return 0, 0, fmt.Errorf("网络请求失败: %v", err)
			}
			time.Sleep(time.Duration(attempt) * time.Second) // 递增等待时间
			continue
		}

		latency = time.Since(startTime).Milliseconds()
		break // 成功，跳出重试循环
	}

	if resp == nil {
		return 0, 0, fmt.Errorf("多次重试后仍然失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 测试下载速度（读取响应body）
	downloadStart := time.Now()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return latency, 0, err
	}
	downloadTime := time.Since(downloadStart)

	// 计算速度 (bytes per second to Mbps)
	bytesPerSecond := float64(len(bodyBytes)) / downloadTime.Seconds()
	mbps := (bytesPerSecond * 8) / (1024 * 1024) // 转换为Mbps

	return latency, mbps, nil
}

// mustParseURL 解析URL，出错时panic
func mustParseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u
}

// isProxyReady 检查代理是否已就绪
func (w *SpeedTestWorkflow) isProxyReady(proxyURL string, timeout time.Duration) bool {
	// 简单检查代理端口是否监听
	u, err := url.Parse(proxyURL)
	if err != nil {
		return false
	}

	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// sortResultsBySpeed 按速度排序结果
func (w *SpeedTestWorkflow) sortResultsBySpeed() {
	sort.Slice(w.results, func(i, j int) bool {
		// 首先按成功与否排序
		if w.results[i].Success != w.results[j].Success {
			return w.results[i].Success
		}

		// 如果都成功，按速度降序排序（快到慢）
		if w.results[i].Success && w.results[j].Success {
			// 如果速度相同，按延迟升序排序
			if w.results[i].Speed == w.results[j].Speed {
				return w.results[i].Latency < w.results[j].Latency
			}
			return w.results[i].Speed > w.results[j].Speed
		}

		// 如果都失败，按节点名称排序
		return w.results[i].Node.Name < w.results[j].Node.Name
	})

	fmt.Printf("📈 结果已按速度排序（从快到慢）\n")
}

// saveResults 保存结果到文件
func (w *SpeedTestWorkflow) saveResults() error {
	file, err := os.Create(w.config.OutputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入标题
	fmt.Fprintf(file, "V2Ray代理节点测速结果\n")
	fmt.Fprintf(file, "测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "订阅链接: %s\n", w.config.SubscriptionURL)
	fmt.Fprintf(file, "测试目标: %s\n", w.config.TestURL)
	fmt.Fprintf(file, "总节点数: %d\n", len(w.results))
	fmt.Fprintf(file, "%s\n", strings.Repeat("=", 80))

	// 统计成功和失败数量
	successCount := 0
	for _, result := range w.results {
		if result.Success {
			successCount++
		}
	}
	fmt.Fprintf(file, "成功节点: %d 个\n", successCount)
	fmt.Fprintf(file, "失败节点: %d 个\n", len(w.results)-successCount)
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("-", 80))

	// 写入成功的节点（按速度排序）
	fmt.Fprintf(file, "📊 成功节点列表（按速度排序：快→慢）\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("-", 80))

	rank := 1
	for _, result := range w.results {
		if result.Success {
			fmt.Fprintf(file, "排名 #%d\n", rank)
			fmt.Fprintf(file, "节点名称: %s\n", result.Node.Name)
			fmt.Fprintf(file, "协议类型: %s\n", result.Node.Protocol)
			fmt.Fprintf(file, "服务器地址: %s:%s\n", result.Node.Server, result.Node.Port)
			fmt.Fprintf(file, "延迟: %d ms\n", result.Latency)
			fmt.Fprintf(file, "下载速度: %.2f Mbps\n", result.Speed)
			fmt.Fprintf(file, "测试时间: %s\n", result.TestTime.Format("15:04:05"))
			fmt.Fprintf(file, "%s\n\n", strings.Repeat("-", 40))
			rank++
		}
	}

	// 写入失败的节点
	fmt.Fprintf(file, "❌ 失败节点列表\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("-", 80))

	for _, result := range w.results {
		if !result.Success {
			fmt.Fprintf(file, "节点名称: %s\n", result.Node.Name)
			fmt.Fprintf(file, "协议类型: %s\n", result.Node.Protocol)
			fmt.Fprintf(file, "服务器地址: %s:%s\n", result.Node.Server, result.Node.Port)
			fmt.Fprintf(file, "失败原因: %s\n", result.Error)
			fmt.Fprintf(file, "测试时间: %s\n", result.TestTime.Format("15:04:05"))
			fmt.Fprintf(file, "%s\n\n", strings.Repeat("-", 40))
		}
	}

	// 同时保存JSON格式的详细结果
	jsonFile := strings.TrimSuffix(w.config.OutputFile, filepath.Ext(w.config.OutputFile)) + ".json"
	jsonData, err := json.MarshalIndent(w.results, "", "  ")
	if err == nil {
		os.WriteFile(jsonFile, jsonData, 0644)
		fmt.Fprintf(file, "\n💾 详细JSON结果已保存到: %s\n", jsonFile)
	}

	fmt.Printf("✅ 结果已保存到: %s\n", w.config.OutputFile)
	if err == nil {
		fmt.Printf("📊 JSON详细结果: %s\n", jsonFile)
	}

	return nil
}

// showSummary 显示测试摘要
func (w *SpeedTestWorkflow) showSummary() {
	fmt.Printf("\n📈 测试摘要:\n")
	fmt.Printf("%s\n", strings.Repeat("=", 50))

	successCount := 0
	totalLatency := int64(0)
	totalSpeed := 0.0
	fastestSpeed := 0.0
	slowestSpeed := float64(^uint(0) >> 1) // 最大float64
	var fastestNode, slowestNode *types.Node

	for _, result := range w.results {
		if result.Success {
			successCount++
			totalLatency += result.Latency
			totalSpeed += result.Speed

			if result.Speed > fastestSpeed {
				fastestSpeed = result.Speed
				fastestNode = result.Node
			}
			if result.Speed < slowestSpeed {
				slowestSpeed = result.Speed
				slowestNode = result.Node
			}
		}
	}

	fmt.Printf("📊 总节点数: %d\n", len(w.results))
	fmt.Printf("✅ 成功节点: %d (%.1f%%)\n", successCount, float64(successCount)/float64(len(w.results))*100)
	fmt.Printf("❌ 失败节点: %d (%.1f%%)\n", len(w.results)-successCount, float64(len(w.results)-successCount)/float64(len(w.results))*100)

	if successCount > 0 {
		fmt.Printf("⚡ 平均延迟: %.1f ms\n", float64(totalLatency)/float64(successCount))
		fmt.Printf("🚀 平均速度: %.2f Mbps\n", totalSpeed/float64(successCount))
		fmt.Printf("🏆 最快节点: %s (%.2f Mbps)\n", fastestNode.Name, fastestSpeed)
		fmt.Printf("🐌 最慢节点: %s (%.2f Mbps)\n", slowestNode.Name, slowestSpeed)
	}

	fmt.Printf("%s\n", strings.Repeat("=", 50))
}

// RunSpeedTestWorkflow 运行测速工作流的入口函数
func RunSpeedTestWorkflow(subscriptionURL string) error {
	workflow := NewSpeedTestWorkflow(subscriptionURL)
	return workflow.Run()
}

// RunCustomSpeedTestWorkflow 运行自定义配置的测速工作流
func RunCustomSpeedTestWorkflow(subscriptionURL string, concurrency int, timeout int, outputFile string, testURL string, maxNodes int) error {
	workflow := NewSpeedTestWorkflow(subscriptionURL)

	if concurrency > 0 {
		workflow.SetConcurrency(concurrency)
	}
	if timeout > 0 {
		workflow.SetTimeout(timeout)
	}
	if outputFile != "" {
		workflow.SetOutputFile(outputFile)
	}
	if testURL != "" {
		workflow.SetTestURL(testURL)
	}
	if maxNodes > 0 {
		workflow.SetMaxNodes(maxNodes)
	}

	return workflow.Run()
}
