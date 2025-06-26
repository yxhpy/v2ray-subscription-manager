package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var proxyManager *ProxyManager
var hysteria2Manager *Hysteria2ProxyManager

func init() {
	proxyManager = NewProxyManager()
	hysteria2Manager = NewHysteria2ProxyManager()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "使用方法: %s <命令> [参数]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n订阅解析命令:\n")
		fmt.Fprintf(os.Stderr, "  parse <订阅链接>                    - 解析订阅链接\n")
		fmt.Fprintf(os.Stderr, "\nV2Ray核心管理:\n")
		fmt.Fprintf(os.Stderr, "  download-v2ray                      - 下载V2Ray核心\n")
		fmt.Fprintf(os.Stderr, "  check-v2ray                         - 检查V2Ray安装状态\n")
		fmt.Fprintf(os.Stderr, "\nHysteria2管理:\n")
		fmt.Fprintf(os.Stderr, "  download-hysteria2                  - 下载Hysteria2客户端\n")
		fmt.Fprintf(os.Stderr, "  check-hysteria2                     - 检查Hysteria2安装状态\n")
		fmt.Fprintf(os.Stderr, "\n代理管理命令:\n")
		fmt.Fprintf(os.Stderr, "  start-proxy random <订阅链接>        - 随机启动代理\n")
		fmt.Fprintf(os.Stderr, "  start-proxy index <订阅链接> <索引>  - 指定节点启动代理\n")
		fmt.Fprintf(os.Stderr, "  start-hysteria2 <订阅链接> <索引>    - 启动Hysteria2代理\n")
		fmt.Fprintf(os.Stderr, "  stop-proxy                          - 停止代理\n")
		fmt.Fprintf(os.Stderr, "  stop-hysteria2                      - 停止Hysteria2代理\n")
		fmt.Fprintf(os.Stderr, "  proxy-status                        - 查看代理状态\n")
		fmt.Fprintf(os.Stderr, "  hysteria2-status                    - 查看Hysteria2状态\n")
		fmt.Fprintf(os.Stderr, "  list-nodes <订阅链接>                - 列出所有节点\n")
		fmt.Fprintf(os.Stderr, "  test-proxy                          - 测试代理连接\n")
		fmt.Fprintf(os.Stderr, "  test-hysteria2                      - 测试Hysteria2连接\n")
		fmt.Fprintf(os.Stderr, "\n测速工作流命令:\n")
		fmt.Fprintf(os.Stderr, "  speed-test <订阅链接>                - 测速工作流(默认配置)\n")
		fmt.Fprintf(os.Stderr, "  speed-test-custom <订阅链接> [选项]   - 自定义测速工作流\n")
		fmt.Fprintf(os.Stderr, "    选项格式: --concurrency=数量 --timeout=秒数 --output=文件名 --test-url=URL\n")
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s parse https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s start-proxy random https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s start-proxy index https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2 5\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s download-v2ray\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s speed-test https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s speed-test-custom https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2 --concurrency=5 --timeout=20\n", os.Args[0])
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "使用方法: %s parse <订阅链接>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "示例: %s parse https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2\n", os.Args[0])
			os.Exit(1)
		}
		if err := ParseSubscription(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "start-proxy":
		handleStartProxy()

	case "stop-proxy":
		if err := proxyManager.StopProxy(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 停止代理失败: %v\n", err)
			os.Exit(1)
		}

	case "proxy-status":
		status := proxyManager.GetStatus()
		statusJSON, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(statusJSON))

		if status.Running {
			fmt.Fprintf(os.Stderr, "✅ 代理运行中\n")
			fmt.Fprintf(os.Stderr, "📡 节点: %s (%s)\n", status.NodeName, status.Protocol)
			fmt.Fprintf(os.Stderr, "🌐 HTTP代理: http://127.0.0.1:%d\n", status.HTTPPort)
			fmt.Fprintf(os.Stderr, "🧦 SOCKS代理: socks5://127.0.0.1:%d\n", status.SOCKSPort)
		} else {
			fmt.Fprintf(os.Stderr, "❌ 代理未运行\n")

			// 检查是否有孤儿进程在运行
			if err := exec.Command("lsof", "-i", ":8080").Run(); err == nil {
				fmt.Fprintf(os.Stderr, "💡 检测到端口8080被占用，可能有V2Ray进程在后台运行\n")
			}
			if err := exec.Command("lsof", "-i", ":1080").Run(); err == nil {
				fmt.Fprintf(os.Stderr, "💡 检测到端口1080被占用，可能有V2Ray进程在后台运行\n")
			}
		}

	case "list-nodes":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "使用方法: %s list-nodes <订阅链接>\n", os.Args[0])
			os.Exit(1)
		}
		nodes, err := getNodesFromSubscription(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 获取节点失败: %v\n", err)
			os.Exit(1)
		}
		ListNodes(nodes)

	case "test-proxy":
		if err := proxyManager.TestProxy(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 代理测试失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "🎉 代理测试通过!\n")

	case "download-v2ray":
		fmt.Println("=== V2Ray核心自动下载器 ===")
		if err := AutoDownloadV2Ray(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 下载安装失败: %v\n", err)
			os.Exit(1)
		}

	case "check-v2ray":
		fmt.Println("=== 检查V2Ray安装状态 ===")
		downloader := NewV2RayDownloader()
		if downloader.CheckV2rayInstalled() {
			fmt.Println("✅ V2Ray已安装")
			downloader.ShowV2rayVersion()
		} else {
			fmt.Println("❌ V2Ray未安装")
			fmt.Printf("运行 '%s download-v2ray' 来安装\n", os.Args[0])
			os.Exit(1)
		}

	case "download-hysteria2":
		fmt.Println("=== Hysteria2客户端自动下载器 ===")
		if err := AutoDownloadHysteria2(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 下载安装失败: %v\n", err)
			os.Exit(1)
		}

	case "check-hysteria2":
		fmt.Println("=== 检查Hysteria2安装状态 ===")
		downloader := NewHysteria2Downloader()
		if downloader.CheckHysteria2Installed() {
			fmt.Println("✅ Hysteria2已安装")
			downloader.ShowHysteria2Version()
		} else {
			fmt.Println("❌ Hysteria2未安装")
			fmt.Printf("运行 '%s download-hysteria2' 来安装\n", os.Args[0])
			os.Exit(1)
		}

	case "start-hysteria2":
		handleStartHysteria2()

	case "stop-hysteria2":
		if err := hysteria2Manager.StopHysteria2Proxy(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 停止Hysteria2代理失败: %v\n", err)
			os.Exit(1)
		}

	case "hysteria2-status":
		status := hysteria2Manager.GetHysteria2Status()
		statusJSON, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(statusJSON))

		if status.Running {
			fmt.Fprintf(os.Stderr, "✅ Hysteria2代理运行中\n")
			fmt.Fprintf(os.Stderr, "📡 节点: %s (%s)\n", status.NodeName, status.Protocol)
			fmt.Fprintf(os.Stderr, "🌐 HTTP代理: http://127.0.0.1:%d\n", status.HTTPPort)
			fmt.Fprintf(os.Stderr, "🧦 SOCKS代理: socks5://127.0.0.1:%d\n", status.SOCKSPort)
		} else {
			fmt.Fprintf(os.Stderr, "❌ Hysteria2代理未运行\n")
		}

	case "test-hysteria2":
		if err := hysteria2Manager.TestHysteria2Proxy(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Hysteria2代理测试失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "🎉 Hysteria2代理测试通过!\n")

	case "speed-test":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "使用方法: %s speed-test <订阅链接>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "示例: %s speed-test https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2\n", os.Args[0])
			os.Exit(1)
		}
		if err := RunSpeedTestWorkflow(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 测速工作流失败: %v\n", err)
			os.Exit(1)
		}

	case "speed-test-custom":
		handleSpeedTestCustom()

	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", command)
		fmt.Fprintf(os.Stderr, "运行 '%s' 不带参数查看可用命令\n", os.Args[0])
		os.Exit(1)
	}
}

// handleStartProxy 处理启动代理命令
func handleStartProxy() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "使用方法:\n")
		fmt.Fprintf(os.Stderr, "  %s start-proxy random <订阅链接>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s start-proxy index <订阅链接> <索引>\n", os.Args[0])
		os.Exit(1)
	}

	mode := os.Args[2]
	subscriptionURL := os.Args[3]

	// 获取节点列表
	nodes, err := getNodesFromSubscription(subscriptionURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 获取节点失败: %v\n", err)
		os.Exit(1)
	}

	switch mode {
	case "random":
		fmt.Fprintf(os.Stderr, "🚀 启动随机代理...\n")
		if err := proxyManager.StartRandomProxy(nodes); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 启动代理失败: %v\n", err)
			os.Exit(1)
		}

	case "index":
		if len(os.Args) != 5 {
			fmt.Fprintf(os.Stderr, "使用方法: %s start-proxy index <订阅链接> <索引>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "提示: 使用 '%s list-nodes <订阅链接>' 查看可用节点\n", os.Args[0])
			os.Exit(1)
		}

		index, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 无效的索引: %s\n", os.Args[4])
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "🚀 启动指定代理...\n")
		if err := proxyManager.StartProxyByIndex(nodes, index); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 启动代理失败: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "❌ 未知模式: %s\n", mode)
		fmt.Fprintf(os.Stderr, "支持的模式: random, index\n")
		os.Exit(1)
	}
}

// handleStartHysteria2 处理启动Hysteria2代理命令
func handleStartHysteria2() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "使用方法: %s start-hysteria2 <订阅链接> <索引>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "提示: 使用 '%s list-nodes <订阅链接>' 查看可用节点\n", os.Args[0])
		os.Exit(1)
	}

	subscriptionURL := os.Args[2]
	index, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 无效的索引: %s\n", os.Args[3])
		os.Exit(1)
	}

	// 获取节点列表
	nodes, err := getNodesFromSubscription(subscriptionURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 获取节点失败: %v\n", err)
		os.Exit(1)
	}

	// 检查索引
	if index < 0 || index >= len(nodes) {
		fmt.Fprintf(os.Stderr, "❌ 索引超出范围: %d (有效范围: 0-%d)\n", index, len(nodes)-1)
		os.Exit(1)
	}

	node := nodes[index]
	if node.Protocol != "hysteria2" {
		fmt.Fprintf(os.Stderr, "❌ 选择的节点不是Hysteria2协议: %s\n", node.Protocol)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "🚀 启动Hysteria2代理...\n")
	fmt.Fprintf(os.Stderr, "📍 选择节点[%d]: %s (%s)\n", index, node.Name, node.Protocol)

	if err := hysteria2Manager.StartHysteria2Proxy(node); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 启动Hysteria2代理失败: %v\n", err)
		os.Exit(1)
	}
}

// getNodesFromSubscription 从订阅链接获取节点列表
func getNodesFromSubscription(subscriptionURL string) ([]*Node, error) {
	// 获取订阅内容
	content, err := fetchSubscription(subscriptionURL)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %v", err)
	}

	// 解码base64
	decoded, err := decodeBase64(content)
	if err != nil {
		return nil, fmt.Errorf("解码失败: %v", err)
	}

	// 解析所有链接
	nodes, err := parseLinks(decoded)
	if err != nil {
		return nil, fmt.Errorf("解析失败: %v", err)
	}

	return nodes, nil
}

// handleSpeedTestCustom 处理自定义测速工作流命令
func handleSpeedTestCustom() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "使用方法: %s speed-test-custom <订阅链接> [选项]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		fmt.Fprintf(os.Stderr, "  --concurrency=数量    并发数 (默认: 10)\n")
		fmt.Fprintf(os.Stderr, "  --timeout=秒数        超时时间 (默认: 30)\n")
		fmt.Fprintf(os.Stderr, "  --output=文件名       输出文件 (默认: speed_test_results.txt)\n")
		fmt.Fprintf(os.Stderr, "  --test-url=URL       测试URL (默认: https://www.google.com)\n")
		fmt.Fprintf(os.Stderr, "  --max-nodes=数量      最大测试节点数 (默认: 不限制)\n")
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s speed-test-custom https://example.com/sub --concurrency=5 --timeout=20\n", os.Args[0])
		os.Exit(1)
	}

	subscriptionURL := os.Args[2]

	// 解析选项
	concurrency := 0
	timeout := 0
	outputFile := ""
	testURL := ""
	maxNodes := 0

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--concurrency=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--concurrency=")); err == nil {
				concurrency = val
			}
		} else if strings.HasPrefix(arg, "--timeout=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--timeout=")); err == nil {
				timeout = val
			}
		} else if strings.HasPrefix(arg, "--output=") {
			outputFile = strings.TrimPrefix(arg, "--output=")
		} else if strings.HasPrefix(arg, "--test-url=") {
			testURL = strings.TrimPrefix(arg, "--test-url=")
		} else if strings.HasPrefix(arg, "--max-nodes=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-nodes=")); err == nil {
				maxNodes = val
			}
		} else {
			fmt.Fprintf(os.Stderr, "未知选项: %s\n", arg)
			os.Exit(1)
		}
	}

	if err := RunCustomSpeedTestWorkflow(subscriptionURL, concurrency, timeout, outputFile, testURL, maxNodes); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 自定义测速工作流失败: %v\n", err)
		os.Exit(1)
	}
}
