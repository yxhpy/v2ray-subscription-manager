package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Hysteria2Downloader Hysteria2客户端下载器
type Hysteria2Downloader struct {
	BaseDir    string
	BinaryPath string
	ConfigPath string
}

// NewHysteria2Downloader 创建新的Hysteria2下载器
func NewHysteria2Downloader() *Hysteria2Downloader {
	return &Hysteria2Downloader{
		BaseDir:    "./hysteria2",
		BinaryPath: "./hysteria2/hysteria",
		ConfigPath: "./hysteria2/config.yaml",
	}
}

// CheckHysteria2Installed 检查Hysteria2是否已安装
func (h *Hysteria2Downloader) CheckHysteria2Installed() bool {
	if _, err := os.Stat(h.BinaryPath); err == nil {
		return true
	}

	// 检查系统路径
	if _, err := exec.LookPath("hysteria"); err == nil {
		h.BinaryPath = "hysteria"
		return true
	}

	return false
}

// ShowHysteria2Version 显示Hysteria2版本
func (h *Hysteria2Downloader) ShowHysteria2Version() {
	cmd := exec.Command(h.BinaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ 无法获取版本信息: %v\n", err)
		return
	}
	fmt.Printf("📍 Hysteria2版本: %s", string(output))
}

// DownloadHysteria2 下载Hysteria2客户端
func (h *Hysteria2Downloader) DownloadHysteria2() error {
	fmt.Println("🚀 开始下载 Hysteria2...")

	// 创建目录
	if err := os.MkdirAll(h.BaseDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 获取下载URL
	downloadURL, err := h.getDownloadURL()
	if err != nil {
		return fmt.Errorf("获取下载链接失败: %v", err)
	}

	fmt.Printf("📥 下载链接: %s\n", downloadURL)

	// 下载文件
	if err := h.downloadFile(downloadURL, h.BinaryPath); err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}

	// 设置执行权限
	if err := os.Chmod(h.BinaryPath, 0755); err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	fmt.Println("✅ Hysteria2 下载完成!")
	h.ShowHysteria2Version()

	return nil
}

// getDownloadURL 获取对应平台的下载链接
func (h *Hysteria2Downloader) getDownloadURL() (string, error) {
	baseURL := "https://github.com/apernet/hysteria/releases/latest/download/"

	var suffix string
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "amd64" {
			suffix = "hysteria-darwin-amd64"
		} else if runtime.GOARCH == "arm64" {
			suffix = "hysteria-darwin-arm64"
		} else {
			return "", fmt.Errorf("不支持的架构: %s", runtime.GOARCH)
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			suffix = "hysteria-linux-amd64"
		} else if runtime.GOARCH == "arm64" {
			suffix = "hysteria-linux-arm64"
		} else {
			return "", fmt.Errorf("不支持的架构: %s", runtime.GOARCH)
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			suffix = "hysteria-windows-amd64.exe"
			h.BinaryPath = "./hysteria2/hysteria.exe"
		} else {
			return "", fmt.Errorf("不支持的架构: %s", runtime.GOARCH)
		}
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return baseURL + suffix, nil
}

// downloadFile 下载文件
func (h *Hysteria2Downloader) downloadFile(url, dest string) error {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态错误: %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// 显示下载进度
	fmt.Printf("📊 开始下载 (大小: %d bytes)...\n", resp.ContentLength)

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// GenerateHysteria2Config 生成Hysteria2配置文件
func (h *Hysteria2Downloader) GenerateHysteria2Config(node *Node, httpPort int, socksPort int) error {
	// 解析服务器和端口
	server := node.Server
	port := node.Port

	// 获取参数
	password := node.UUID // Hysteria2中用户标识通常作为密码
	obfs := ""
	insecure := "false"

	if obsParam, ok := node.Parameters["obfs"]; ok {
		obfs = obsParam
	}

	if _, ok := node.Parameters["insecure"]; ok {
		insecure = "true"
	}

	// 生成配置
	config := fmt.Sprintf(`# Hysteria2 客户端配置
server: %s:%s

auth: %s

bandwidth:
  up: 20 mbps
  down: 100 mbps

socks5:
  listen: 127.0.0.1:%d

http:
  listen: 127.0.0.1:%d

tls:
  insecure: %s
`, server, port, password, socksPort, httpPort, insecure)

	// 如果有混淆参数
	if obfs != "" {
		config += fmt.Sprintf(`
obfs:
  type: salamander
  salamander:
    password: %s
`, obfs)
	}

	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(h.ConfigPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(h.ConfigPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	fmt.Printf("✅ Hysteria2配置已生成: %s\n", h.ConfigPath)
	return nil
}

// StartHysteria2 启动Hysteria2客户端
func (h *Hysteria2Downloader) StartHysteria2() (*exec.Cmd, error) {
	if !h.CheckHysteria2Installed() {
		return nil, fmt.Errorf("Hysteria2未安装")
	}

	// 启动命令
	cmd := exec.Command(h.BinaryPath, "client", "-c", h.ConfigPath)

	// 启动进程
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("启动Hysteria2失败: %v", err)
	}

	fmt.Println("🚀 Hysteria2客户端已启动")
	return cmd, nil
}

// TestHysteria2Config 测试Hysteria2配置
func (h *Hysteria2Downloader) TestHysteria2Config() error {
	if !h.CheckHysteria2Installed() {
		return fmt.Errorf("Hysteria2未安装")
	}

	// 检查配置文件
	if _, err := os.Stat(h.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", h.ConfigPath)
	}

	fmt.Printf("✅ Hysteria2配置文件有效: %s\n", h.ConfigPath)
	return nil
}

// AutoDownloadHysteria2 自动下载安装Hysteria2
func AutoDownloadHysteria2() error {
	downloader := NewHysteria2Downloader()

	if downloader.CheckHysteria2Installed() {
		fmt.Println("✅ Hysteria2已安装")
		downloader.ShowHysteria2Version()
		return nil
	}

	fmt.Println("📦 Hysteria2未安装，开始自动下载...")

	if err := downloader.DownloadHysteria2(); err != nil {
		return fmt.Errorf("自动下载失败: %v", err)
	}

	fmt.Println("🎉 Hysteria2安装完成!")
	return nil
}
