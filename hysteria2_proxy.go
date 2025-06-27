package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Hysteria2ProxyManager Hysteria2代理管理器
type Hysteria2ProxyManager struct {
	Downloader  *Hysteria2Downloader
	Process     *exec.Cmd
	CurrentNode *Node
	HTTPPort    int
	SOCKSPort   int
}

// NewHysteria2ProxyManager 创建新的Hysteria2代理管理器
func NewHysteria2ProxyManager() *Hysteria2ProxyManager {
	downloader := NewHysteria2Downloader()
	// 为每个实例生成唯一的配置文件路径
	downloader.ConfigPath = fmt.Sprintf("./hysteria2/config_%d.yaml", time.Now().UnixNano())

	return &Hysteria2ProxyManager{
		Downloader: downloader,
		HTTPPort:   8081, // 使用不同端口避免冲突
		SOCKSPort:  1081,
	}
}

// StartHysteria2Proxy 启动Hysteria2代理
func (h *Hysteria2ProxyManager) StartHysteria2Proxy(node *Node) error {
	if node.Protocol != "hysteria2" {
		return fmt.Errorf("节点协议不是Hysteria2: %s", node.Protocol)
	}

	// 检查Hysteria2是否安装
	if !h.Downloader.CheckHysteria2Installed() {
		fmt.Println("🔽 Hysteria2未安装，正在自动下载...")
		if err := h.Downloader.SafeDownloadHysteria2(); err != nil {
			return fmt.Errorf("自动下载Hysteria2失败: %v", err)
		}
	}

	// 停止现有代理
	if h.Process != nil {
		h.StopHysteria2Proxy()
	}

	// 分配端口（如果还未设置）
	if h.HTTPPort == 0 || h.HTTPPort == 8081 {
		h.HTTPPort = findAvailablePort(8081)
	}
	if h.SOCKSPort == 0 || h.SOCKSPort == 1081 {
		h.SOCKSPort = findAvailablePort(1081)
	}

	fmt.Printf("🔧 配置代理端口: HTTP=%d, SOCKS=%d\n", h.HTTPPort, h.SOCKSPort)

	// 生成配置文件
	if err := h.Downloader.GenerateHysteria2Config(node, h.HTTPPort, h.SOCKSPort); err != nil {
		return fmt.Errorf("生成配置失败: %v", err)
	}

	// 启动Hysteria2客户端
	process, err := h.Downloader.StartHysteria2()
	if err != nil {
		return fmt.Errorf("启动Hysteria2失败: %v", err)
	}

	h.Process = process
	h.CurrentNode = node

	// 等待启动
	fmt.Println("⏳ 等待Hysteria2启动...")
	time.Sleep(3 * time.Second)

	// 检查是否成功启动
	if !h.IsHysteria2Running() {
		h.Process = nil
		h.CurrentNode = nil
		return fmt.Errorf("Hysteria2启动失败或意外退出")
	}

	fmt.Printf("✅ Hysteria2代理启动成功!\n")
	fmt.Printf("📡 节点: %s\n", node.Name)
	fmt.Printf("🌐 HTTP代理: http://127.0.0.1:%d\n", h.HTTPPort)
	fmt.Printf("🧦 SOCKS代理: socks5://127.0.0.1:%d\n", h.SOCKSPort)

	return nil
}

// StopHysteria2Proxy 停止Hysteria2代理
func (h *Hysteria2ProxyManager) StopHysteria2Proxy() error {
	if h.Process == nil {
		return fmt.Errorf("没有运行中的Hysteria2代理")
	}

	// 发送终止信号
	if h.Process.Process != nil {
		err := h.Process.Process.Signal(syscall.SIGTERM)
		if err != nil {
			// 如果温和终止失败，强制杀死
			h.Process.Process.Kill()
		}
	}

	// 等待进程结束
	h.Process.Wait()
	h.Process = nil
	h.CurrentNode = nil

	// 清理临时配置文件
	if h.Downloader != nil && h.Downloader.ConfigPath != "./hysteria2/config.yaml" {
		os.Remove(h.Downloader.ConfigPath)
	}

	fmt.Println("🛑 Hysteria2代理已停止")
	return nil
}

// IsHysteria2Running 检查Hysteria2是否运行
func (h *Hysteria2ProxyManager) IsHysteria2Running() bool {
	// 首先检查进程状态
	if h.Process != nil && h.Process.Process != nil {
		err := h.Process.Process.Signal(syscall.Signal(0))
		if err == nil {
			return true
		}
	}

	// 通过端口检查
	if h.HTTPPort > 0 && h.SOCKSPort > 0 {
		// 检查HTTP端口
		httpConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", h.HTTPPort), 1*time.Second)
		if err == nil {
			httpConn.Close()
			// 检查SOCKS端口
			socksConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", h.SOCKSPort), 1*time.Second)
			if err == nil {
				socksConn.Close()
				return true
			}
		}
	}

	return false
}

// TestHysteria2Proxy 测试Hysteria2代理连接
func (h *Hysteria2ProxyManager) TestHysteria2Proxy() error {
	if !h.IsHysteria2Running() {
		return fmt.Errorf("Hysteria2代理未运行")
	}

	// 测试HTTP代理
	fmt.Println("🧪 测试Hysteria2 HTTP代理连接...")
	httpConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", h.HTTPPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("HTTP代理连接失败: %v", err)
	}
	httpConn.Close()
	fmt.Println("✅ HTTP代理连接正常")

	// 测试SOCKS代理
	fmt.Println("🧪 测试Hysteria2 SOCKS代理连接...")
	socksConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", h.SOCKSPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("SOCKS代理连接失败: %v", err)
	}
	socksConn.Close()
	fmt.Println("✅ SOCKS代理连接正常")

	return nil
}

// GetHysteria2Status 获取Hysteria2代理状态
func (h *Hysteria2ProxyManager) GetHysteria2Status() ProxyStatus {
	status := ProxyStatus{
		Running:   h.IsHysteria2Running(),
		HTTPPort:  h.HTTPPort,
		SOCKSPort: h.SOCKSPort,
	}

	if h.CurrentNode != nil {
		status.NodeName = h.CurrentNode.Name
		status.Protocol = h.CurrentNode.Protocol
		status.Server = h.CurrentNode.Server
	}

	return status
}
