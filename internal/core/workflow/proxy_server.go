package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// ProxyServer 代理服务器
type ProxyServer struct {
	configFile       string
	httpPort         int
	socksPort        int
	currentNode      *types.ValidNode
	proxyManager     *proxy.ProxyManager
	hysteria2Manager *proxy.Hysteria2ProxyManager
	mutex            sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	watcher          *fsnotify.Watcher
}

// NewProxyServer 创建新的代理服务器
func NewProxyServer(configFile string, httpPort, socksPort int) *ProxyServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &ProxyServer{
		configFile:       configFile,
		httpPort:         httpPort,
		socksPort:        socksPort,
		proxyManager:     proxy.NewProxyManager(),
		hysteria2Manager: proxy.NewHysteria2ProxyManager(),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start 启动代理服务器
func (ps *ProxyServer) Start() error {
	fmt.Printf("🚀 启动代理服务器...\n")
	fmt.Printf("📁 配置文件: %s\n", ps.configFile)
	fmt.Printf("🌐 HTTP端口: %d\n", ps.httpPort)
	fmt.Printf("🧦 SOCKS端口: %d\n", ps.socksPort)

	// 设置信号处理
	ps.setupSignalHandler()

	// 启动文件监控（无论文件是否存在）
	if err := ps.startFileWatcher(); err != nil {
		return fmt.Errorf("启动文件监控失败: %v", err)
	}

	// 尝试加载初始配置
	if err := ps.loadConfig(); err != nil {
		fmt.Printf("⚠️ 初始配置加载失败: %v\n", err)
		fmt.Printf("⏳ 等待配置文件出现...\n")
	} else {
		// 启动初始代理
		if err := ps.startProxy(); err != nil {
			fmt.Printf("⚠️ 启动初始代理失败: %v\n", err)
			fmt.Printf("⏳ 等待有效配置...\n")
		} else {
			fmt.Printf("✅ 代理服务器启动成功！\n")
			fmt.Printf("🌐 HTTP代理: http://127.0.0.1:%d\n", ps.httpPort)
			fmt.Printf("🧦 SOCKS代理: socks5://127.0.0.1:%d\n", ps.socksPort)
		}
	}

	fmt.Printf("👁️ 监控配置文件变化中...\n")
	fmt.Printf("📝 按 Ctrl+C 停止服务\n")

	// 阻塞等待
	<-ps.ctx.Done()
	return nil
}

// Stop 停止代理服务器
func (ps *ProxyServer) Stop() error {
	fmt.Printf("🛑 停止代理服务器...\n")

	ps.cancel()

	// 停止文件监控
	if ps.watcher != nil {
		fmt.Printf("  🛑 停止文件监控...\n")
		ps.watcher.Close()
	}

	// 停止代理
	fmt.Printf("  🛑 停止代理进程...\n")
	ps.stopProxy()

	// 清理临时配置文件
	fmt.Printf("  🧹 清理临时配置文件...\n")
	ps.cleanupTempFiles()

	// 杀死相关进程
	fmt.Printf("  💀 杀死相关进程...\n")
	ps.killRelatedProcesses()

	fmt.Printf("✅ 代理服务器已完全停止\n")
	return nil
}

// loadConfig 加载配置文件
func (ps *ProxyServer) loadConfig() error {
	fmt.Printf("📄 加载配置文件: %s\n", ps.configFile)

	data, err := os.ReadFile(ps.configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var state MVPState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	if state.BestNode == nil {
		return fmt.Errorf("配置文件中没有最佳节点信息")
	}

	ps.mutex.Lock()
	ps.currentNode = state.BestNode
	ps.mutex.Unlock()

	fmt.Printf("✅ 配置加载成功\n")
	fmt.Printf("📡 当前节点: %s (%s)\n", ps.currentNode.Node.Name, ps.currentNode.Node.Protocol)
	fmt.Printf("📊 节点性能: 延迟 %dms, 速度 %.2f Mbps, 分数 %.2f\n",
		ps.currentNode.Latency, ps.currentNode.Speed, ps.currentNode.Score)

	return nil
}

// startProxy 启动代理
func (ps *ProxyServer) startProxy() error {
	ps.mutex.RLock()
	node := ps.currentNode
	ps.mutex.RUnlock()

	if node == nil {
		return fmt.Errorf("没有可用的节点配置")
	}

	fmt.Printf("🚀 启动代理: %s (%s)\n", node.Node.Name, node.Node.Protocol)

	// 停止现有代理
	ps.stopProxy()

	switch node.Node.Protocol {
	case "vmess", "vless", "trojan", "ss":
		return ps.startV2RayProxy(node.Node)
	case "hysteria2":
		return ps.startHysteria2Proxy(node.Node)
	default:
		return fmt.Errorf("不支持的协议: %s", node.Node.Protocol)
	}
}

// startV2RayProxy 启动V2Ray代理
func (ps *ProxyServer) startV2RayProxy(node *types.Node) error {
	// 设置固定端口
	ps.proxyManager.HTTPPort = ps.httpPort
	ps.proxyManager.SOCKSPort = ps.socksPort
	ps.proxyManager.ConfigPath = fmt.Sprintf("proxy_server_v2ray_%d.json", time.Now().UnixNano())

	err := ps.proxyManager.StartProxy(node)
	if err != nil {
		return fmt.Errorf("启动V2Ray代理失败: %v", err)
	}

	fmt.Printf("✅ V2Ray代理启动成功\n")
	return nil
}

// startHysteria2Proxy 启动Hysteria2代理
func (ps *ProxyServer) startHysteria2Proxy(node *types.Node) error {
	// 设置固定端口
	ps.hysteria2Manager.HTTPPort = ps.httpPort
	ps.hysteria2Manager.SOCKSPort = ps.socksPort
	ps.hysteria2Manager.SetConfigPath(fmt.Sprintf("./hysteria2/proxy_server_%d.yaml", time.Now().UnixNano()))

	err := ps.hysteria2Manager.StartHysteria2Proxy(node)
	if err != nil {
		return fmt.Errorf("启动Hysteria2代理失败: %v", err)
	}

	fmt.Printf("✅ Hysteria2代理启动成功\n")
	return nil
}

// stopProxy 停止代理
func (ps *ProxyServer) stopProxy() {
	if ps.proxyManager != nil {
		ps.proxyManager.StopProxy()
	}
	if ps.hysteria2Manager != nil {
		ps.hysteria2Manager.StopHysteria2Proxy()
	}
}

// startFileWatcher 启动文件监控
func (ps *ProxyServer) startFileWatcher() error {
	var err error
	ps.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %v", err)
	}

	// 尝试监控配置文件，如果文件不存在则监控当前目录
	err = ps.watcher.Add(ps.configFile)
	if err != nil {
		// 如果文件不存在，监控当前目录来检测文件创建
		fmt.Printf("📁 配置文件不存在，监控当前目录等待文件创建\n")
		err = ps.watcher.Add(".")
		if err != nil {
			return fmt.Errorf("添加目录监控失败: %v", err)
		}
	}

	// 启动监控协程
	go ps.watchFileChanges()

	return nil
}

// watchFileChanges 监控文件变化
func (ps *ProxyServer) watchFileChanges() {
	for {
		select {
		case event, ok := <-ps.watcher.Events:
			if !ok {
				return
			}

			// 检查是否是我们关心的配置文件
			if event.Name == ps.configFile {
				// 处理写入和创建事件
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("📝 检测到配置文件变化: %s\n", event.Name)
					ps.handleConfigChange()
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("📄 检测到配置文件创建: %s\n", event.Name)
					// 文件创建后，尝试添加直接监控
					ps.watcher.Remove(".")
					if err := ps.watcher.Add(ps.configFile); err == nil {
						fmt.Printf("✅ 已切换到直接监控配置文件\n")
					}
					ps.handleConfigChange()
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Printf("🗑️ 检测到配置文件被删除: %s\n", event.Name)
					fmt.Printf("⏳ 继续使用当前节点，等待配置文件恢复...\n")
					// 切换回监控目录
					ps.watcher.Remove(ps.configFile)
					ps.watcher.Add(".")
				}
			}

		case err, ok := <-ps.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("⚠️ 文件监控错误: %v\n", err)

		case <-ps.ctx.Done():
			return
		}
	}
}

// handleConfigChange 处理配置文件变化
func (ps *ProxyServer) handleConfigChange() {
	// 等待一下，确保文件写入完成
	time.Sleep(1 * time.Second)

	fmt.Printf("🔄 处理配置变化...\n")

	// 加载新配置
	data, err := os.ReadFile(ps.configFile)
	if err != nil {
		fmt.Printf("❌ 读取新配置失败: %v\n", err)
		return
	}

	var state MVPState
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("❌ 解析新配置失败: %v\n", err)
		return
	}

	if state.BestNode == nil {
		fmt.Printf("❌ 新配置中没有最佳节点信息\n")
		return
	}

	// 检查是否需要切换
	ps.mutex.RLock()
	currentNode := ps.currentNode
	ps.mutex.RUnlock()

	newNode := state.BestNode

	// 如果是同一个节点，不需要切换
	if currentNode != nil &&
		currentNode.Node.Name == newNode.Node.Name &&
		currentNode.Node.Server == newNode.Node.Server &&
		currentNode.Node.Port == newNode.Node.Port {
		fmt.Printf("📊 节点未变化，无需切换\n")
		return
	}

	fmt.Printf("🔍 发现新节点，开始测试...\n")
	fmt.Printf("📡 新节点: %s (分数: %.2f)\n", newNode.Node.Name, newNode.Score)
	if currentNode != nil {
		fmt.Printf("📡 当前节点: %s (分数: %.2f)\n", currentNode.Node.Name, currentNode.Score)
	}

	// 测试新节点
	if ps.testNode(newNode.Node) {
		fmt.Printf("✅ 新节点测试通过，开始切换...\n")

		ps.mutex.Lock()
		ps.currentNode = newNode
		ps.mutex.Unlock()

		if err := ps.startProxy(); err != nil {
			fmt.Printf("❌ 切换到新节点失败: %v\n", err)
			// 回滚到原节点
			ps.mutex.Lock()
			ps.currentNode = currentNode
			ps.mutex.Unlock()
			ps.startProxy()
		} else {
			fmt.Printf("🎉 成功切换到新节点: %s\n", newNode.Node.Name)
		}
	} else {
		fmt.Printf("❌ 新节点测试失败，保持当前节点\n")
	}
}

// testNode 测试节点连通性
func (ps *ProxyServer) testNode(node *types.Node) bool {
	fmt.Printf("🧪 测试节点: %s (%s)\n", node.Name, node.Protocol)

	// 使用临时端口测试
	testHTTPPort := ps.httpPort + 1000
	testSOCKSPort := ps.socksPort + 1000

	var err error

	switch node.Protocol {
	case "vmess", "vless", "trojan", "ss":
		v2rayMgr := proxy.NewProxyManager()
		v2rayMgr.HTTPPort = testHTTPPort
		v2rayMgr.SOCKSPort = testSOCKSPort
		v2rayMgr.ConfigPath = fmt.Sprintf("test_proxy_%d.json", time.Now().UnixNano())

		err = v2rayMgr.StartProxy(node)
		defer v2rayMgr.StopProxy()

	case "hysteria2":
		hysteria2Mgr := proxy.NewHysteria2ProxyManager()
		hysteria2Mgr.HTTPPort = testHTTPPort
		hysteria2Mgr.SOCKSPort = testSOCKSPort
		hysteria2Mgr.SetConfigPath(fmt.Sprintf("./hysteria2/test_proxy_%d.yaml", time.Now().UnixNano()))

		err = hysteria2Mgr.StartHysteria2Proxy(node)
		defer hysteria2Mgr.StopHysteria2Proxy()

	default:
		fmt.Printf("❌ 不支持的协议: %s\n", node.Protocol)
		return false
	}

	if err != nil {
		fmt.Printf("❌ 启动测试代理失败: %v\n", err)
		return false
	}

	// 等待代理启动
	time.Sleep(3 * time.Second)

	// 执行详细的连通性测试
	success := ps.detailedConnectivityTest(testHTTPPort)

	if success {
		fmt.Printf("✅ 节点测试通过\n")
	} else {
		fmt.Printf("❌ 节点测试失败\n")
	}

	return success
}

// simpleConnectivityTest 简单的连通性测试
func (ps *ProxyServer) simpleConnectivityTest(httpPort int) bool {
	// 这里可以实现一个简单的HTTP请求测试
	// 为了简化，我们假设如果代理能启动就认为测试通过
	// 在实际应用中，可以发送HTTP请求来验证连通性
	return true
}

// detailedConnectivityTest 详细的连通性测试
func (ps *ProxyServer) detailedConnectivityTest(httpPort int) bool {
	// 创建代理客户端
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		fmt.Printf("❌ 解析代理URL失败: %v\n", err)
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURLParsed),
		},
		Timeout: 15 * time.Second,
	}

	// 尝试多个测试URL
	testURLs := []string{
		"http://httpbin.org/ip",
		"http://www.google.com",
		"http://www.baidu.com",
	}

	for _, testURL := range testURLs {
		resp, err := client.Get(testURL)
		if err != nil {
			fmt.Printf("🔍 测试URL %s 失败: %v\n", testURL, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			// 读取响应内容以确保连接正常
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil && len(body) > 0 {
				fmt.Printf("✅ 连通性测试通过 - URL: %s, 响应大小: %d bytes\n", testURL, len(body))
				return true
			}
		}
		resp.Body.Close()
	}

	fmt.Printf("❌ 所有测试URL都失败\n")
	return false
}

// setupSignalHandler 设置信号处理
func (ps *ProxyServer) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("\n🛑 接收到退出信号，正在停止服务...\n")
		ps.Stop()
		os.Exit(0)
	}()
}

// cleanupTempFiles 清理临时文件
func (ps *ProxyServer) cleanupTempFiles() {
	patterns := []string{
		"proxy_server_v2ray_*.json",
		"proxy_server_hysteria2_*.json",
		"temp_v2ray_config_*.json",
		"temp_hysteria2_config_*.json",
		"test_proxy_*.json",
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

	// 清理hysteria2目录下的临时文件
	hysteria2Patterns := []string{
		"./hysteria2/proxy_server_*.yaml",
		"./hysteria2/temp_*.yaml",
		"./hysteria2/test_proxy_*.yaml",
	}

	for _, pattern := range hysteria2Patterns {
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
func (ps *ProxyServer) killRelatedProcesses() {
	processNames := []string{"v2ray", "xray", "hysteria2", "hysteria"}

	for _, processName := range processNames {
		cmd := exec.Command("pkill", "-f", processName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("    💀 已终止 %s 进程\n", processName)
		}
	}

	// 特别处理占用端口的进程
	ports := []int{ps.httpPort, ps.socksPort}
	for _, port := range ports {
		if pid := ps.getProcessByPort(port); pid > 0 {
			cmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
			if err := cmd.Run(); err == nil {
				fmt.Printf("    💀 已终止占用端口 %d 的进程 (PID: %d)\n", port, pid)
			}
		}
	}
}

// getProcessByPort 获取占用指定端口的进程PID
func (ps *ProxyServer) getProcessByPort(port int) int {
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0
	}

	var pid int
	if _, err := fmt.Sscanf(pidStr, "%d", &pid); err == nil && pid > 0 {
		return pid
	}
	return 0
}

// RunProxyServer 运行代理服务器
func RunProxyServer(configFile string, httpPort, socksPort int) error {
	server := NewProxyServer(configFile, httpPort, socksPort)
	return server.Start()
}

// RunDualProxySystem 运行双进程代理系统
func RunDualProxySystem(subscriptionURL string, httpPort, socksPort int) error {
	fmt.Printf("🚀 启动双进程代理系统...\n")
	fmt.Printf("📡 订阅链接: %s\n", subscriptionURL)
	fmt.Printf("🌐 HTTP端口: %d\n", httpPort)
	fmt.Printf("🧦 SOCKS端口: %d\n", socksPort)

	// 状态文件路径
	stateFile := "mvp_best_node.json"

	// 创建MVP测试器
	tester := NewMVPTester(subscriptionURL)
	tester.SetStateFile(stateFile)
	tester.SetInterval(5 * time.Minute) // 每5分钟测试一次
	tester.SetMaxNodes(50)              // 最多测试50个节点
	tester.SetConcurrency(5)            // 并发数为5

	// 创建代理服务器
	server := NewProxyServer(stateFile, httpPort, socksPort)

	// 设置信号处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 启动MVP测试器
	go func() {
		if err := tester.Start(); err != nil {
			fmt.Printf("❌ MVP测试器启动失败: %v\n", err)
		}
	}()

	// 等待一下，让测试器先运行
	time.Sleep(2 * time.Second)

	// 启动代理服务器
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("❌ 代理服务器启动失败: %v\n", err)
		}
	}()

	fmt.Printf("✅ 双进程代理系统启动成功！\n")
	fmt.Printf("📝 按 Ctrl+C 停止服务\n")

	// 等待停止信号
	<-c
	fmt.Printf("\n🛑 接收到停止信号，正在停止系统...\n")

	// 停止所有服务
	fmt.Printf("  🛑 停止MVP测试器...\n")
	tester.Stop()

	fmt.Printf("  🛑 停止代理服务器...\n")
	server.Stop()

	fmt.Printf("✅ 双进程代理系统已完全停止\n")
	return nil
}
