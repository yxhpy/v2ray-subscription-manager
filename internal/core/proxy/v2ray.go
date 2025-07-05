package proxy

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/platform"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// 全局互斥锁用于端口分配
var portAllocationMutex sync.Mutex

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	Running   bool   `json:"running"`
	HTTPPort  int    `json:"http_port"`
	SOCKSPort int    `json:"socks_port"`
	NodeName  string `json:"node_name"`
	Protocol  string `json:"protocol"`
	Server    string `json:"server"`
}

// ProxyManager V2Ray代理管理器
type ProxyManager struct {
	ConfigPath   string
	HTTPPort     int
	SOCKSPort    int
	V2RayProcess *exec.Cmd
	CurrentNode  *types.Node
}

// ProxyState 代理状态持久化结构
type ProxyState struct {
	HTTPPort    int    `json:"http_port"`
	SOCKSPort   int    `json:"socks_port"`
	NodeName    string `json:"node_name"`
	Protocol    string `json:"protocol"`
	Server      string `json:"server"`
	ConfigPath  string `json:"config_path"`
	LastUpdated int64  `json:"last_updated"`
}

const StateFile = "proxy_state.json"

// NewProxyManager 创建新的代理管理器
func NewProxyManager() *ProxyManager {
	// 为每个实例生成唯一的配置文件路径，避免并发冲突
	uniqueConfigPath := fmt.Sprintf("temp_v2ray_config_%d.json", time.Now().UnixNano())

	pm := &ProxyManager{
		ConfigPath: uniqueConfigPath,
		HTTPPort:   0, // 将自动分配
		SOCKSPort:  0, // 将自动分配
	}

	// 注意：不加载状态，确保每个管理器实例都是独立的
	// 这样每次都会分配新的端口，避免端口冲突

	fmt.Printf("DEBUG: 创建新的代理管理器，配置文件: %s\n", uniqueConfigPath)

	return pm
}

// NewTestProxyManager 创建用于测试的代理管理器（不加载状态）
func NewTestProxyManager() *ProxyManager {
	// 为测试实例生成带测试前缀的唯一配置文件路径
	uniqueConfigPath := fmt.Sprintf("test_v2ray_config_%d_%d.json", time.Now().UnixNano(), os.Getpid())

	pm := &ProxyManager{
		ConfigPath: uniqueConfigPath,
		HTTPPort:   0, // 将自动分配
		SOCKSPort:  0, // 将自动分配
	}

	// 测试实例不加载状态，保持完全独立
	return pm
}

// 全局端口计数器，确保每次分配不同的端口
var globalPortCounter int64

// 已占用端口记录
var usedPorts = make(map[int]bool)
var usedPortsMutex sync.RWMutex

// findAvailablePort 查找可用端口（改进版本，确保端口唯一）
func findAvailablePort(startPort int) int {
	fmt.Printf("DEBUG: 开始查找可用端口，起始端口: %d\n", startPort)

	portAllocationMutex.Lock()
	defer portAllocationMutex.Unlock()

	fmt.Printf("DEBUG: 获得端口分配锁\n")

	// 使用全局计数器确保每次分配不同的起始点
	globalPortCounter++

	// 从多个不同的范围开始搜索，避免集中在一个区域
	searchRanges := []int{
		8000 + int(globalPortCounter*37%500), // 8000-8499
		8500 + int(globalPortCounter*41%500), // 8500-8999
		9000 + int(globalPortCounter*43%500), // 9000-9499
		9500 + int(globalPortCounter*47%500), // 9500-9999
	}

	// 添加随机偏移
	randomOffset := rand.Intn(50)

	fmt.Printf("DEBUG: 搜索范围起始点: %v, 计数器: %d, 随机偏移: %d\n", searchRanges, globalPortCounter, randomOffset)

	// 在每个范围内搜索可用端口
	for _, basePort := range searchRanges {
		for i := 0; i < 50; i++ {
			candidatePort := basePort + i + randomOffset

			// 确保端口在合理范围内
			if candidatePort < 8000 {
				candidatePort = 8000 + (candidatePort % 100)
			}
			if candidatePort > 65535 {
				candidatePort = 8000 + (candidatePort % 2000)
			}

			// 检查是否已被我们记录为占用
			usedPortsMutex.RLock()
			alreadyUsed := usedPorts[candidatePort]
			usedPortsMutex.RUnlock()

			if alreadyUsed {
				fmt.Printf("DEBUG: 端口 %d 已被记录为占用，跳过\n", candidatePort)
				continue
			}

			// 使用TCP服务器测试端口可用性
			if isPortAvailable(candidatePort) {
				// 记录端口为已使用
				usedPortsMutex.Lock()
				usedPorts[candidatePort] = true
				usedPortsMutex.Unlock()

				fmt.Printf("DEBUG: 找到并分配可用端口: %d\n", candidatePort)
				return candidatePort
			}
		}
	}

	// 最后备选方案：使用计数器直接生成一个端口
	fallbackPort := 8000 + int(globalPortCounter%2000)
	fmt.Printf("DEBUG: 使用fallback端口: %d\n", fallbackPort)

	// 记录fallback端口
	usedPortsMutex.Lock()
	usedPorts[fallbackPort] = true
	usedPortsMutex.Unlock()

	return fallbackPort
}

// releasePort 释放端口（当连接关闭时调用）
func releasePort(port int) {
	usedPortsMutex.Lock()
	defer usedPortsMutex.Unlock()
	delete(usedPorts, port)
	fmt.Printf("DEBUG: 释放端口: %d\n", port)
}

// isPortAvailable 检查端口是否可用（通过建立TCP服务器测试）
func isPortAvailable(port int) bool {
	// 方法1：尝试监听该端口
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		// 监听失败说明端口被占用
		fmt.Printf("DEBUG: 端口 %d 被占用 (监听失败: %v)\n", port, err)
		return false
	}
	listener.Close()

	// 方法2：尝试连接该端口确认没有服务在运行
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
	if err == nil {
		// 如果连接成功，说明端口被占用
		conn.Close()
		fmt.Printf("DEBUG: 端口 %d 被占用 (连接成功)\n", port)
		return false
	}

	fmt.Printf("DEBUG: 端口 %d 可用\n", port)
	return true
}

// generateV2RayConfig 生成V2Ray配置
func generateV2RayConfig(node *types.Node, httpPort, socksPort int) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "http",
				"port":     httpPort,
				"protocol": "http",
				"listen":   "0.0.0.0",
				"settings": map[string]interface{}{
					"allowTransparent": false,
					"timeout":         300,
				},
				"sockopt": map[string]interface{}{
					"tcpKeepAlive": true,
					"reusePort":    true,
				},
			},
			{
				"tag":      "socks",
				"port":     socksPort,
				"protocol": "socks",
				"listen":   "0.0.0.0",
				"settings": map[string]interface{}{
					"auth": "noauth",
					"udp":  true,
					// 关键修复：移除限制性的IP参数，确保可以接受局域网连接
				},
				"sockopt": map[string]interface{}{
					"tcpKeepAlive": true,
					"reusePort":    true,
				},
			},
		},
		"routing": map[string]interface{}{
			"domainStrategy": "IPOnDemand",
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"inboundTag":  []string{"http", "socks"},
					"outboundTag": "proxy",
				},
				// 允许局域网访问，不阻止私有IP段
			},
		},
		"outbounds": []map[string]interface{}{},
	}

	var outbound map[string]interface{}

	switch node.Protocol {
	case "vless":
		outbound = map[string]interface{}{
			"tag":      "proxy",
			"protocol": "vless",
			"settings": map[string]interface{}{
				"vnext": []map[string]interface{}{
					{
						"address": node.Server,
						"port":    parsePort(node.Port),
						"users": []map[string]interface{}{
							{
								"id":         node.UUID,
								"encryption": "none",
							},
						},
					},
				},
			},
			"streamSettings": generateStreamSettings(node),
		}

	case "ss":
		// 检查加密方法兼容性
		supportedMethods := map[string]string{
			// 传统方法 - 转换为更安全的AEAD方法
			"aes-256-cfb":   "aes-256-gcm",
			"aes-128-cfb":   "aes-128-gcm",
			"aes-192-cfb":   "aes-256-gcm",
			"aes-256-ctr":   "aes-256-gcm",
			"aes-128-ctr":   "aes-128-gcm",
			"chacha20":      "chacha20-poly1305",
			"chacha20-ietf": "chacha20-poly1305",

			// AEAD方法 - 直接支持
			"aes-256-gcm":             "aes-256-gcm",
			"aes-128-gcm":             "aes-128-gcm",
			"chacha20-poly1305":       "chacha20-poly1305",
			"chacha20-ietf-poly1305":  "chacha20-poly1305", // 常见的chacha20变体
			"xchacha20-poly1305":      "chacha20-poly1305", // 扩展版本，映射到标准版本
			"xchacha20-ietf-poly1305": "chacha20-poly1305", // 扩展版本变体

			// Shadowsocks 2022新方法
			"2022-blake3-aes-128-gcm":       "2022-blake3-aes-128-gcm",
			"2022-blake3-aes-256-gcm":       "2022-blake3-aes-256-gcm",
			"2022-blake3-chacha20-poly1305": "2022-blake3-chacha20-poly1305",
			"2022-blake3-chacha12-poly1305": "2022-blake3-chacha12-poly1305",
			"2022-blake3-chacha8-poly1305":  "2022-blake3-chacha8-poly1305",

			// 无加密方法（通常与TLS结合使用）
			"none":  "none",
			"plain": "none",
		}

		method := node.Method
		if convertedMethod, ok := supportedMethods[method]; ok {
			method = convertedMethod
			if method != node.Method {
				fmt.Fprintf(os.Stderr, "警告: 加密方法 %s 已转换为 %s\n", node.Method, method)
			}
		} else {
			fmt.Fprintf(os.Stderr, "警告: 不支持的加密方法 %s，使用默认方法 aes-256-gcm\n", node.Method)
			method = "aes-256-gcm"
		}

		// Shadowsocks配置 - 针对Windows环境优化
		servers := []map[string]interface{}{
			{
				"address":  node.Server,
				"port":     parsePort(node.Port),
				"method":   method,
				"password": node.Password,
			},
		}

		outbound = map[string]interface{}{
			"tag":      "proxy",
			"protocol": "shadowsocks",
			"settings": map[string]interface{}{
				"servers": servers,
			},
			"streamSettings": map[string]interface{}{
				"network": "tcp",
				"tcpSettings": map[string]interface{}{
					"header": map[string]interface{}{
						"type": "none",
					},
				},
				"sockopt": map[string]interface{}{
					"tcpKeepAliveInterval": 30,   // TCP Keep-Alive间隔
					"tcpNoDelay":           true, // 禁用Nagle算法，提高响应速度
				},
			},
		}

	case "vmess":
		outbound = map[string]interface{}{
			"tag":      "proxy",
			"protocol": "vmess",
			"settings": map[string]interface{}{
				"vnext": []map[string]interface{}{
					{
						"address": node.Server,
						"port":    parsePort(node.Port),
						"users": []map[string]interface{}{
							{
								"id":       node.UUID,
								"alterId":  parseAlterId(node.Parameters["aid"]),
								"security": getVmessSecurity(node.Parameters["scy"]),
							},
						},
					},
				},
			},
			"streamSettings": generateVmessStreamSettings(node),
		}

	case "trojan":
		outbound = map[string]interface{}{
			"tag":      "proxy",
			"protocol": "trojan",
			"settings": map[string]interface{}{
				"servers": []map[string]interface{}{
					{
						"address":  node.Server,
						"port":     parsePort(node.Port),
						"password": node.Password,
					},
				},
			},
			"streamSettings": generateTrojanStreamSettings(node),
		}

	case "hysteria2":
		// V2Ray不直接支持Hysteria2，这里提供一个基本的转换
		// 实际使用中可能需要其他工具
		fmt.Fprintf(os.Stderr, "警告: V2Ray不直接支持Hysteria2协议，跳过此节点\n")
		return nil, fmt.Errorf("V2Ray不支持Hysteria2协议")

	default:
		return nil, fmt.Errorf("不支持的协议类型: %s", node.Protocol)
	}

	// 添加直连出站
	directOutbound := map[string]interface{}{
		"tag":      "direct",
		"protocol": "freedom",
		"settings": map[string]interface{}{},
	}

	config["outbounds"] = []map[string]interface{}{outbound, directOutbound}

	return config, nil
}

// generateStreamSettings 生成流设置
func generateStreamSettings(node *types.Node) map[string]interface{} {
	streamSettings := map[string]interface{}{
		"network": "tcp", // 默认TCP
	}

	// 检查传输类型
	if transportType, ok := node.Parameters["type"]; ok {
		switch transportType {
		case "ws":
			streamSettings["network"] = "ws"
			wsSettings := map[string]interface{}{}

			if path, ok := node.Parameters["path"]; ok && path != "" {
				wsSettings["path"] = path
			}

			if host, ok := node.Parameters["host"]; ok && host != "" {
				wsSettings["headers"] = map[string]interface{}{
					"Host": host,
				}
			}

			streamSettings["wsSettings"] = wsSettings

		case "grpc":
			streamSettings["network"] = "grpc"
			grpcSettings := map[string]interface{}{}

			if serviceName, ok := node.Parameters["serviceName"]; ok && serviceName != "" {
				grpcSettings["serviceName"] = serviceName
			}

			streamSettings["grpcSettings"] = grpcSettings

		case "h2":
			streamSettings["network"] = "h2"
			h2Settings := map[string]interface{}{}

			if path, ok := node.Parameters["path"]; ok && path != "" {
				h2Settings["path"] = path
			}

			if host, ok := node.Parameters["host"]; ok && host != "" {
				h2Settings["host"] = []string{host}
			}

			streamSettings["httpSettings"] = h2Settings
		}
	}

	// TLS设置
	if security, ok := node.Parameters["security"]; ok && security == "tls" {
		streamSettings["security"] = "tls"
		tlsSettings := map[string]interface{}{}

		if sni, ok := node.Parameters["sni"]; ok && sni != "" {
			tlsSettings["serverName"] = sni
		}

		if insecure, ok := node.Parameters["allowInsecure"]; ok && insecure == "1" {
			tlsSettings["allowInsecure"] = true
		}

		if fp, ok := node.Parameters["fp"]; ok && fp != "" {
			tlsSettings["fingerprint"] = fp
		}

		streamSettings["tlsSettings"] = tlsSettings
	}

	// TCP头部设置
	if headerType, ok := node.Parameters["headerType"]; ok {
		streamSettings["tcpSettings"] = map[string]interface{}{
			"header": map[string]interface{}{
				"type": headerType,
			},
		}
	}

	return streamSettings
}

// parsePort 解析端口号
func parsePort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 443 // 默认端口
	}
	return port
}

// parseAlterId 解析VMess alterId
func parseAlterId(aidStr string) int {
	if aidStr == "" {
		return 0 // V2Ray 4.28+ 默认为0
	}
	aid, err := strconv.Atoi(aidStr)
	if err != nil {
		return 0
	}
	return aid
}

// getVmessSecurity 获取VMess加密方式
func getVmessSecurity(scy string) string {
	switch scy {
	case "auto", "aes-128-gcm", "chacha20-poly1305", "none":
		return scy
	default:
		return "auto" // 默认auto
	}
}

// generateVmessStreamSettings 生成VMess流设置
func generateVmessStreamSettings(node *types.Node) map[string]interface{} {
	streamSettings := map[string]interface{}{
		"network": "tcp", // 默认TCP
	}

	// 检查网络类型
	if net, ok := node.Parameters["net"]; ok {
		streamSettings["network"] = net

		switch net {
		case "ws":
			wsSettings := map[string]interface{}{}

			if path, ok := node.Parameters["path"]; ok && path != "" {
				wsSettings["path"] = path
			}

			if host, ok := node.Parameters["host"]; ok && host != "" {
				wsSettings["headers"] = map[string]interface{}{
					"Host": host,
				}
			}

			streamSettings["wsSettings"] = wsSettings

		case "h2":
			h2Settings := map[string]interface{}{}

			if path, ok := node.Parameters["path"]; ok && path != "" {
				h2Settings["path"] = path
			}

			if host, ok := node.Parameters["host"]; ok && host != "" {
				h2Settings["host"] = []string{host}
			}

			streamSettings["httpSettings"] = h2Settings

		case "grpc":
			grpcSettings := map[string]interface{}{}

			if serviceName, ok := node.Parameters["serviceName"]; ok && serviceName != "" {
				grpcSettings["serviceName"] = serviceName
			}

			streamSettings["grpcSettings"] = grpcSettings
		}
	}

	// TLS设置
	if tls, ok := node.Parameters["tls"]; ok && tls == "tls" {
		streamSettings["security"] = "tls"
		tlsSettings := map[string]interface{}{}

		if sni, ok := node.Parameters["sni"]; ok && sni != "" {
			tlsSettings["serverName"] = sni
		}

		if fp, ok := node.Parameters["fp"]; ok && fp != "" {
			tlsSettings["fingerprint"] = fp
		}

		if alpn, ok := node.Parameters["alpn"]; ok && alpn != "" {
			tlsSettings["alpn"] = []string{alpn}
		}

		streamSettings["tlsSettings"] = tlsSettings
	}

	return streamSettings
}

// generateTrojanStreamSettings 生成Trojan流设置
func generateTrojanStreamSettings(node *types.Node) map[string]interface{} {
	streamSettings := map[string]interface{}{
		"network": "tcp", // 默认TCP
	}

	// Trojan通常使用TLS
	streamSettings["security"] = "tls"
	tlsSettings := map[string]interface{}{}

	// 从参数中提取TLS设置
	if sni, ok := node.Parameters["sni"]; ok && sni != "" {
		tlsSettings["serverName"] = sni
	}

	if insecure, ok := node.Parameters["allowInsecure"]; ok && insecure == "1" {
		tlsSettings["allowInsecure"] = true
	}

	if fp, ok := node.Parameters["fp"]; ok && fp != "" {
		tlsSettings["fingerprint"] = fp
	}

	if alpn, ok := node.Parameters["alpn"]; ok && alpn != "" {
		tlsSettings["alpn"] = []string{alpn}
	}

	streamSettings["tlsSettings"] = tlsSettings

	return streamSettings
}

// StartRandomProxy 启动随机代理
func (pm *ProxyManager) StartRandomProxy(nodes []*types.Node) error {
	if len(nodes) == 0 {
		return fmt.Errorf("没有可用的节点")
	}

	// 过滤支持的协议
	supportedNodes := []*types.Node{}
	for _, node := range nodes {
		if node.Protocol == "vless" || node.Protocol == "ss" || node.Protocol == "vmess" || node.Protocol == "trojan" {
			supportedNodes = append(supportedNodes, node)
		}
	}

	if len(supportedNodes) == 0 {
		return fmt.Errorf("没有支持的协议节点 (支持VLESS、SS、VMess、Trojan)")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(supportedNodes))
	selectedNode := supportedNodes[randomIndex]

	// 找到原始索引用于显示
	originalIndex := -1
	for i, node := range nodes {
		if node == selectedNode {
			originalIndex = i
			break
		}
	}

	fmt.Fprintf(os.Stderr, "🎲 随机选择节点[%d]: %s (%s)\n", originalIndex, selectedNode.Name, selectedNode.Protocol)
	return pm.StartProxy(selectedNode)
}

// StartProxyByIndex 按索引启动代理
func (pm *ProxyManager) StartProxyByIndex(nodes []*types.Node, index int) error {
	if index < 0 || index >= len(nodes) {
		return fmt.Errorf("节点索引 %d 超出范围 (0-%d)", index, len(nodes)-1)
	}

	selectedNode := nodes[index]
	fmt.Fprintf(os.Stderr, "📍 选择节点[%d]: %s (%s)\n", index, selectedNode.Name, selectedNode.Protocol)
	return pm.StartProxy(selectedNode)
}

// StartProxy 启动代理
func (pm *ProxyManager) StartProxy(node *types.Node) error {
	// 检查V2Ray是否安装
	if !pm.checkV2RayInstalled() {
		return fmt.Errorf("V2Ray未安装，请先运行: %s download-v2ray", os.Args[0])
	}

	// 停止现有代理
	if pm.V2RayProcess != nil {
		pm.StopProxy()
	}

	// 分配端口（只在端口为0时才重新分配）
	if pm.HTTPPort == 0 {
		pm.HTTPPort = findAvailablePort(8080)
	}
	if pm.SOCKSPort == 0 {
		pm.SOCKSPort = findAvailablePort(1080)
	}

	fmt.Fprintf(os.Stderr, "🔧 配置代理端口: HTTP=%d, SOCKS=%d\n", pm.HTTPPort, pm.SOCKSPort)

	// 生成配置
	config, err := generateV2RayConfig(node, pm.HTTPPort, pm.SOCKSPort)
	if err != nil {
		return fmt.Errorf("生成配置失败: %v", err)
	}

	// 保存配置文件
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	err = os.WriteFile(pm.ConfigPath, configJSON, 0644)
	if err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	// 启动V2Ray
	v2rayPath := "./v2ray/v2ray"
	if runtime.GOOS == "windows" {
		v2rayPath = "./v2ray/v2ray.exe"
	}

	if _, err := os.Stat(v2rayPath); os.IsNotExist(err) {
		v2rayPath = "v2ray" // 尝试系统路径
	}

	pm.V2RayProcess = exec.Command(v2rayPath, "run", "-c", pm.ConfigPath)
	pm.CurrentNode = node

	// 设置进程组，便于管理
	platform.SetProcAttributes(pm.V2RayProcess)

	err = pm.V2RayProcess.Start()
	if err != nil {
		pm.V2RayProcess = nil
		pm.CurrentNode = nil
		return fmt.Errorf("启动V2Ray失败: %v", err)
	}

	// 等待一下确保启动成功
	time.Sleep(2 * time.Second)

	// 检查进程是否仍在运行
	if !pm.isV2RayRunning() {
		pm.V2RayProcess = nil
		pm.CurrentNode = nil
		return fmt.Errorf("V2Ray进程启动后意外退出，可能是配置问题")
	}

	fmt.Fprintf(os.Stderr, "✅ 代理启动成功!\n")
	fmt.Fprintf(os.Stderr, "📡 节点: %s\n", node.Name)
	fmt.Fprintf(os.Stderr, "🌐 HTTP代理: http://127.0.0.1:%d\n", pm.HTTPPort)
	fmt.Fprintf(os.Stderr, "🧦 SOCKS代理: socks5://127.0.0.1:%d\n", pm.SOCKSPort)

	// 保存状态
	pm.saveState()

	return nil
}

// StopProxy 停止代理
func (pm *ProxyManager) StopProxy() error {
	if pm.V2RayProcess == nil {
		return fmt.Errorf("没有运行中的代理")
	}

	// 记录要释放的端口
	httpPortToRelease := pm.HTTPPort
	socksPortToRelease := pm.SOCKSPort

	// 发送终止信号
	if pm.V2RayProcess.Process != nil {
		err := pm.V2RayProcess.Process.Signal(syscall.SIGTERM)
		if err != nil {
			// 如果温和终止失败，强制杀死
			pm.V2RayProcess.Process.Kill()
		}
	}

	// 等待进程结束
	pm.V2RayProcess.Wait()
	pm.V2RayProcess = nil
	pm.CurrentNode = nil

	// 释放端口资源
	if httpPortToRelease > 0 {
		releasePort(httpPortToRelease)
	}
	if socksPortToRelease > 0 && socksPortToRelease != httpPortToRelease {
		releasePort(socksPortToRelease)
	}

	// 重置端口
	pm.HTTPPort = 0
	pm.SOCKSPort = 0

	// 清理配置文件
	if _, err := os.Stat(pm.ConfigPath); err == nil {
		os.Remove(pm.ConfigPath)
	}

	// 清除状态
	pm.saveState()

	fmt.Fprintf(os.Stderr, "🛑 代理已停止\n")
	return nil
}

// GetStatus 获取代理状态
func (pm *ProxyManager) GetStatus() ProxyStatus {
	status := ProxyStatus{
		Running:   pm.isV2RayRunning(),
		HTTPPort:  pm.HTTPPort,
		SOCKSPort: pm.SOCKSPort,
	}

	if pm.CurrentNode != nil {
		status.NodeName = pm.CurrentNode.Name
		status.Protocol = pm.CurrentNode.Protocol
		status.Server = pm.CurrentNode.Server
	}

	return status
}

// isV2RayRunning 检查V2Ray进程是否运行
func (pm *ProxyManager) isV2RayRunning() bool {
	// 首先检查保存的进程状态
	if pm.V2RayProcess != nil && pm.V2RayProcess.Process != nil {
		err := pm.V2RayProcess.Process.Signal(syscall.Signal(0))
		if err == nil {
			return true
		}
	}

	// 如果进程对象检查失败，则通过端口检查
	if pm.HTTPPort > 0 && pm.SOCKSPort > 0 {
		// 检查HTTP端口
		httpConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", pm.HTTPPort), 1*time.Second)
		if err == nil {
			httpConn.Close()
			// 检查SOCKS端口
			socksConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", pm.SOCKSPort), 1*time.Second)
			if err == nil {
				socksConn.Close()
				return true
			}
		}
	}

	return false
}

// IsRunning 检查代理是否正在运行（公开方法）
func (pm *ProxyManager) IsRunning() bool {
	return pm.isV2RayRunning()
}

// GetCurrentNode 获取当前连接的节点
func (pm *ProxyManager) GetCurrentNode() *types.Node {
	return pm.CurrentNode
}

// SetFixedPorts 设置固定端口
func (pm *ProxyManager) SetFixedPorts(httpPort, socksPort int) {
	pm.HTTPPort = httpPort
	pm.SOCKSPort = socksPort
}

// IsPortOccupied 检查指定端口是否被当前代理占用
func (pm *ProxyManager) IsPortOccupied(port int) bool {
	if !pm.isV2RayRunning() {
		return false
	}
	return pm.HTTPPort == port || pm.SOCKSPort == port
}

// checkV2RayInstalled 检查V2Ray是否安装
func (pm *ProxyManager) checkV2RayInstalled() bool {
	paths := []string{"./v2ray/v2ray", "v2ray"}
	if runtime.GOOS == "windows" {
		paths = []string{"./v2ray/v2ray.exe", "v2ray"}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		if _, err := exec.LookPath(path); err == nil {
			return true
		}
	}
	return false
}

// ListNodes 列出所有节点（带索引）
func ListNodes(nodes []*types.Node) {
	fmt.Fprintf(os.Stderr, "\n📋 可用节点列表:\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for i, node := range nodes {
		fmt.Fprintf(os.Stderr, "[%3d] %-10s %s\n", i, fmt.Sprintf("(%s)", node.Protocol), node.Name)
	}

	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "总计: %d 个节点\n\n", len(nodes))
}

// TestProxy 测试代理连接
func (pm *ProxyManager) TestProxy() error {
	if !pm.isV2RayRunning() {
		return fmt.Errorf("代理未运行")
	}

	// 测试HTTP代理
	fmt.Fprintf(os.Stderr, "🧪 测试HTTP代理连接...\n")
	httpConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", pm.HTTPPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("HTTP代理连接失败: %v", err)
	}
	httpConn.Close()
	fmt.Fprintf(os.Stderr, "✅ HTTP代理连接正常\n")

	// 测试SOCKS代理
	fmt.Fprintf(os.Stderr, "🧪 测试SOCKS代理连接...\n")
	socksConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", pm.SOCKSPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("SOCKS代理连接失败: %v", err)
	}
	socksConn.Close()
	fmt.Fprintf(os.Stderr, "✅ SOCKS代理连接正常\n")

	return nil
}

// saveState 保存代理状态
func (pm *ProxyManager) saveState() error {
	if pm.CurrentNode == nil {
		// 删除状态文件
		os.Remove(StateFile)
		return nil
	}

	state := ProxyState{
		HTTPPort:    pm.HTTPPort,
		SOCKSPort:   pm.SOCKSPort,
		NodeName:    pm.CurrentNode.Name,
		Protocol:    pm.CurrentNode.Protocol,
		Server:      pm.CurrentNode.Server,
		ConfigPath:  pm.ConfigPath,
		LastUpdated: time.Now().Unix(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(StateFile, data, 0644)
}

// loadState 加载代理状态
func (pm *ProxyManager) loadState() {
	data, err := os.ReadFile(StateFile)
	if err != nil {
		return // 文件不存在或读取失败，忽略
	}

	var state ProxyState
	if err := json.Unmarshal(data, &state); err != nil {
		return // 解析失败，忽略
	}

	// 检查状态是否过期（超过1小时）
	if time.Now().Unix()-state.LastUpdated > 3600 {
		os.Remove(StateFile)
		return
	}

	// 恢复状态
	pm.HTTPPort = state.HTTPPort
	pm.SOCKSPort = state.SOCKSPort
	pm.ConfigPath = state.ConfigPath

	// 创建虚拟types.Node对象
	pm.CurrentNode = &types.Node{
		Name:     state.NodeName,
		Protocol: state.Protocol,
		Server:   state.Server,
	}
}
