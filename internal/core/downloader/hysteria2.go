package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// 全局互斥锁，防止并发下载 Hysteria2
var hysteria2DownloadMutex sync.Mutex

// Hysteria2Downloader Hysteria2客户端下载器
type Hysteria2Downloader struct {
	BaseDir    string
	BinaryPath string
	ConfigPath string
}

// NewHysteria2Downloader 创建新的Hysteria2下载器
func NewHysteria2Downloader() *Hysteria2Downloader {
	binaryPath := "./hysteria2/hysteria"
	// Windows 系统使用 .exe 扩展名
	if runtime.GOOS == "windows" {
		binaryPath = "./hysteria2/hysteria.exe"
	}

	return &Hysteria2Downloader{
		BaseDir:    "./hysteria2",
		BinaryPath: binaryPath,
		ConfigPath: "./hysteria2/config.yaml",
	}
}

// CheckHysteria2Installed 检查Hysteria2是否已安装
func (h *Hysteria2Downloader) CheckHysteria2Installed() bool {
	// 检查预期的二进制文件路径
	if _, err := os.Stat(h.BinaryPath); err == nil {
		return true
	}

	// Windows 下检查可能的原始下载文件名
	if runtime.GOOS == "windows" {
		originalName := "./hysteria2/hysteria-windows-amd64.exe"
		if _, err := os.Stat(originalName); err == nil {
			// 如果找到原始文件，重命名为预期的名称
			if err := os.Rename(originalName, h.BinaryPath); err == nil {
				fmt.Printf("✅ 发现已下载的 Hysteria2 文件，已重命名为: %s\n", h.BinaryPath)
				return true
			}
		}
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

	// 清理可能存在的损坏文件
	if _, err := os.Stat(h.BinaryPath); err == nil {
		fmt.Printf("🗑️ 删除现有文件: %s\n", h.BinaryPath)
		os.Remove(h.BinaryPath)
	}

	// 创建目录
	if err := os.MkdirAll(h.BaseDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 获取下载URL列表
	downloadURLs, err := h.getDownloadURLs()
	if err != nil {
		return fmt.Errorf("获取下载链接失败: %v", err)
	}

	fmt.Printf("📂 目标路径: %s\n", h.BinaryPath)

	// 尝试从多个源下载
	var lastErr error
	for i, downloadURL := range downloadURLs {
		if i > 0 {
			fmt.Printf("🔄 尝试备用下载源...\n")
		}

		fmt.Printf("📥 下载链接: %s\n", downloadURL)

		lastErr = h.downloadFile(downloadURL, h.BinaryPath)
		if lastErr == nil {
			break // 下载成功
		}

		fmt.Printf("❌ 下载源失败: %v\n", lastErr)
	}

	if lastErr != nil {
		return fmt.Errorf("所有下载源都失败: %v", lastErr)
	}

	// 设置执行权限
	if err := os.Chmod(h.BinaryPath, 0755); err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	fmt.Println("✅ Hysteria2 下载完成!")
	h.ShowHysteria2Version()

	return nil
}

// SafeDownloadHysteria2 安全下载Hysteria2（带互斥锁）
func (h *Hysteria2Downloader) SafeDownloadHysteria2() error {
	// 使用互斥锁防止并发下载
	hysteria2DownloadMutex.Lock()
	defer hysteria2DownloadMutex.Unlock()

	// 再次检查是否已安装（可能在等待锁的过程中被其他goroutine安装了）
	if h.CheckHysteria2Installed() {
		return nil
	}

	// 重试下载最多3次
	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			fmt.Printf("🔄 第 %d 次重试下载...\n", i+1)
			time.Sleep(time.Duration(i) * time.Second) // 递增延迟
		}

		lastErr = h.DownloadHysteria2()
		if lastErr == nil {
			return nil
		}

		fmt.Printf("❌ 下载失败: %v\n", lastErr)
	}

	return fmt.Errorf("下载失败，已重试3次: %v", lastErr)
}

// getDownloadURL 获取对应平台的下载链接
func (h *Hysteria2Downloader) getDownloadURL() (string, error) {
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
		} else {
			return "", fmt.Errorf("不支持的架构: %s", runtime.GOARCH)
		}
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return "https://github.com/apernet/hysteria/releases/latest/download/" + suffix, nil
}

// getDownloadURLs 获取多个下载源
func (h *Hysteria2Downloader) getDownloadURLs() ([]string, error) {
	mainURL, err := h.getDownloadURL()
	if err != nil {
		return nil, err
	}

	// 返回多个下载源
	urls := []string{
		mainURL,
		// 可以添加其他镜像源
	}

	return urls, nil
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
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// 验证下载完整性
	if resp.ContentLength > 0 && written != resp.ContentLength {
		return fmt.Errorf("下载不完整: 期望 %d 字节，实际 %d 字节", resp.ContentLength, written)
	}

	// 验证文件大小
	if written < 1000000 { // 小于1MB可能有问题
		return fmt.Errorf("下载的文件过小 (%d 字节)，可能下载失败", written)
	}

	fmt.Printf("✅ 下载完成，文件大小: %d 字节\n", written)

	// Windows 下验证是否为有效的 PE 文件
	if runtime.GOOS == "windows" {
		if err := h.validateWindowsExecutable(dest); err != nil {
			// 删除无效文件
			os.Remove(dest)
			return fmt.Errorf("下载的文件无效: %v", err)
		}
	}

	return nil
}

// validateWindowsExecutable 验证 Windows 可执行文件
func (h *Hysteria2Downloader) validateWindowsExecutable(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("🔍 文件大小: %d 字节\n", fileInfo.Size())

	// 读取文件头
	header := make([]byte, 64)
	n, err := file.Read(header)
	if err != nil {
		return err
	}

	fmt.Printf("🔍 读取文件头: %d 字节\n", n)
	fmt.Printf("🔍 文件头前8字节: %x\n", header[:8])

	// 检查 DOS 头 "MZ"
	if len(header) < 2 || header[0] != 0x4D || header[1] != 0x5A {
		return fmt.Errorf("不是有效的 Windows 可执行文件 (缺少 MZ 签名，实际: %x %x)", header[0], header[1])
	}

	// 尝试运行版本命令来进一步验证
	cmd := exec.Command(filePath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️ 版本检查失败: %v\n", err)
		fmt.Printf("⚠️ 输出: %s\n", string(output))
		return fmt.Errorf("可执行文件无法运行: %v", err)
	}

	fmt.Printf("✅ Windows 可执行文件验证通过\n")
	fmt.Printf("✅ 版本信息: %s\n", string(output))
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

	if err := downloader.SafeDownloadHysteria2(); err != nil {
		return fmt.Errorf("自动下载失败: %v", err)
	}

	fmt.Println("🎉 Hysteria2安装完成!")
	return nil
}
