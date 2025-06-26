package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// V2RayDownloader V2Ray下载器结构体
type V2RayDownloader struct {
	Version     string
	InstallPath string
	TempDir     string
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS   string
	Arch string
}

// DownloadMirror 下载镜像源
type DownloadMirror struct {
	Name string
	URL  string
}

// NewV2RayDownloader 创建新的下载器实例
func NewV2RayDownloader() *V2RayDownloader {
	return &V2RayDownloader{
		Version:     "latest", // 可以指定版本，如 "v5.12.1"
		InstallPath: "./v2ray",
		TempDir:     "./temp",
	}
}

// GetSystemInfo 获取系统信息
func (d *V2RayDownloader) GetSystemInfo() SystemInfo {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 标准化操作系统名称
	switch osName {
	case "darwin":
		osName = "macos" // V2Ray使用macos而不是darwin
	}

	// 标准化架构名称以匹配V2Ray发布文件命名
	switch arch {
	case "amd64":
		arch = "64"
	case "386":
		arch = "32"
	case "arm64":
		// 根据实际的V2Ray发布文件名格式
		if runtime.GOOS == "darwin" {
			arch = "arm64" // macOS ARM64: v2ray-macos-arm64.zip
		} else if osName == "android" {
			arch = "arm64-v8a" // Android ARM64: v2ray-android-arm64-v8a.zip
		} else {
			arch = "arm64" // Linux ARM64: v2ray-linux-arm64.zip
		}
	case "arm":
		if osName == "android" {
			arch = "arm32-v7a"
		} else {
			arch = "arm32"
		}
	}

	return SystemInfo{
		OS:   osName,
		Arch: arch,
	}
}

// CheckV2rayInstalled 检查V2Ray是否已安装
func (d *V2RayDownloader) CheckV2rayInstalled() bool {
	// 检查本地安装路径
	v2rayPath := filepath.Join(d.InstallPath, d.getV2rayExecutableName())
	if _, err := os.Stat(v2rayPath); err == nil {
		fmt.Printf("在本地路径找到V2Ray: %s\n", v2rayPath)
		return true
	}

	// 检查系统PATH中的v2ray
	cmd := exec.Command("v2ray", "-version")
	if err := cmd.Run(); err == nil {
		fmt.Println("在系统PATH中找到V2Ray")
		return true
	}

	fmt.Println("未找到V2Ray核心，需要下载安装")
	return false
}

// getV2rayExecutableName 获取V2Ray可执行文件名
func (d *V2RayDownloader) getV2rayExecutableName() string {
	if runtime.GOOS == "windows" {
		return "v2ray.exe"
	}
	return "v2ray"
}

// GetDownloadMirrors 获取下载镜像源列表
func (d *V2RayDownloader) GetDownloadMirrors(sysInfo SystemInfo) []DownloadMirror {
	version := d.Version
	if version == "latest" {
		version = "v5.33.0" // 使用最新稳定版本
	}

	// 尝试多种可能的文件名格式
	possibleFileNames := d.getPossibleFileNames(sysInfo)

	var mirrors []DownloadMirror

	// 为每种可能的文件名创建镜像源
	for _, fileName := range possibleFileNames {
		mirrors = append(mirrors, []DownloadMirror{
			{
				Name: fmt.Sprintf("GitHub Official (%s)", fileName),
				URL:  fmt.Sprintf("https://github.com/v2fly/v2ray-core/releases/download/%s/%s", version, fileName),
			},
			{
				Name: fmt.Sprintf("GitHub Mirror (%s)", fileName),
				URL:  fmt.Sprintf("https://ghproxy.com/https://github.com/v2fly/v2ray-core/releases/download/%s/%s", version, fileName),
			},
		}...)
	}

	return mirrors
}

// getPossibleFileNames 获取可能的文件名列表
func (d *V2RayDownloader) getPossibleFileNames(sysInfo SystemInfo) []string {
	var fileNames []string

	// 为macOS尝试多种文件名格式
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			fileNames = append(fileNames,
				"v2ray-macos-arm64.zip",
				"v2ray-darwin-arm64.zip",
				"v2ray-macos-arm64-v8a.zip",
				"v2ray-darwin-arm64-v8a.zip",
			)
		} else {
			fileNames = append(fileNames,
				"v2ray-macos-64.zip",
				"v2ray-darwin-64.zip",
			)
		}
	} else {
		// 其他系统使用标准格式
		fileName := fmt.Sprintf("v2ray-%s-%s.zip", sysInfo.OS, sysInfo.Arch)
		fileNames = append(fileNames, fileName)
	}

	return fileNames
}

// DownloadWithProgress 带进度显示的文件下载
func (d *V2RayDownloader) DownloadWithProgress(url, fileName string) error {
	// 创建临时目录
	if err := os.MkdirAll(d.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}

	filePath := filepath.Join(d.TempDir, fileName)

	// 创建文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	// 发起HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 获取文件大小
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	// 进度跟踪
	var downloaded int64 = 0
	buffer := make([]byte, 32*1024) // 32KB buffer

	fmt.Printf("开始下载: %s\n", fileName)
	fmt.Printf("文件大小: %.2f MB\n", float64(size)/1024/1024)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(n)

			// 显示进度
			if size > 0 {
				progress := float64(downloaded) / float64(size) * 100
				fmt.Printf("\r下载进度: %.1f%% (%.2f/%.2f MB)",
					progress,
					float64(downloaded)/1024/1024,
					float64(size)/1024/1024)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %v", err)
		}
	}

	fmt.Println("\n下载完成!")
	return nil
}

// ExtractZip 解压ZIP文件
func (d *V2RayDownloader) ExtractZip(src, dest string) error {
	// 打开ZIP文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开ZIP文件失败: %v", err)
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	fmt.Println("正在解压文件...")

	// 解压文件
	for _, f := range r.File {
		// 构建文件路径
		path := filepath.Join(dest, f.Name)

		// 确保路径安全
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			// 创建目录
			os.MkdirAll(path, f.FileInfo().Mode())
			continue
		}

		// 创建文件的目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("创建文件目录失败: %v", err)
		}

		// 打开ZIP中的文件
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("打开ZIP内文件失败: %v", err)
		}

		// 创建目标文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("创建目标文件失败: %v", err)
		}

		// 复制文件内容
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("复制文件内容失败: %v", err)
		}

		fmt.Printf("解压: %s\n", f.Name)
	}

	fmt.Println("解压完成!")
	return nil
}

// SetExecutablePermission 设置可执行权限（Linux/macOS）
func (d *V2RayDownloader) SetExecutablePermission(filePath string) error {
	if runtime.GOOS == "windows" {
		return nil // Windows不需要设置执行权限
	}

	fmt.Printf("设置执行权限: %s\n", filePath)
	return os.Chmod(filePath, 0755)
}

// CleanupTempFiles 清理临时文件
func (d *V2RayDownloader) CleanupTempFiles() {
	if err := os.RemoveAll(d.TempDir); err != nil {
		fmt.Printf("清理临时文件失败: %v\n", err)
	} else {
		fmt.Println("清理临时文件完成")
	}
}

// DownloadAndInstall 下载并安装V2Ray
func (d *V2RayDownloader) DownloadAndInstall() error {
	// 检查是否已安装
	if d.CheckV2rayInstalled() {
		return nil
	}

	// 获取系统信息
	sysInfo := d.GetSystemInfo()
	fmt.Printf("检测到系统: %s-%s\n", sysInfo.OS, sysInfo.Arch)

	// 获取下载镜像源
	mirrors := d.GetDownloadMirrors(sysInfo)

	var downloadErr error
	var downloadSuccess bool

	// 尝试从不同镜像源下载
	for _, mirror := range mirrors {
		fmt.Printf("\n尝试从 %s 下载...\n", mirror.Name)
		fmt.Printf("下载链接: %s\n", mirror.URL)

		fileName := filepath.Base(mirror.URL)

		// 下载文件
		downloadErr = d.DownloadWithProgress(mirror.URL, fileName)
		if downloadErr != nil {
			fmt.Printf("从 %s 下载失败: %v\n", mirror.Name, downloadErr)
			continue
		}

		downloadSuccess = true

		// 解压文件
		zipPath := filepath.Join(d.TempDir, fileName)
		extractErr := d.ExtractZip(zipPath, d.InstallPath)
		if extractErr != nil {
			fmt.Printf("解压失败: %v\n", extractErr)
			downloadSuccess = false
			continue
		}

		// 设置执行权限
		v2rayPath := filepath.Join(d.InstallPath, d.getV2rayExecutableName())
		if err := d.SetExecutablePermission(v2rayPath); err != nil {
			fmt.Printf("设置执行权限失败: %v\n", err)
			downloadSuccess = false
			continue
		}

		break
	}

	// 清理临时文件
	d.CleanupTempFiles()

	if !downloadSuccess {
		// 提供手动下载指导
		d.ShowManualDownloadGuide(sysInfo)
		return fmt.Errorf("从所有镜像源下载都失败，最后一个错误: %v", downloadErr)
	}

	// 验证安装
	v2rayPath := filepath.Join(d.InstallPath, d.getV2rayExecutableName())
	if _, err := os.Stat(v2rayPath); err != nil {
		return fmt.Errorf("安装验证失败，找不到V2Ray可执行文件: %s", v2rayPath)
	}

	fmt.Printf("\n✅ V2Ray核心安装成功！\n")
	fmt.Printf("安装路径: %s\n", v2rayPath)

	// 显示版本信息
	d.ShowV2rayVersion()

	return nil
}

// ShowV2rayVersion 显示V2Ray版本信息
func (d *V2RayDownloader) ShowV2rayVersion() {
	v2rayPath := filepath.Join(d.InstallPath, d.getV2rayExecutableName())
	cmd := exec.Command(v2rayPath, "-version")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("无法获取V2Ray版本信息: %v\n", err)
		return
	}

	fmt.Printf("V2Ray版本信息:\n%s\n", string(output))
}

// ShowManualDownloadGuide 显示手动下载指导
func (d *V2RayDownloader) ShowManualDownloadGuide(sysInfo SystemInfo) {
	fmt.Println("\n📋 手动下载指南:")
	fmt.Println("由于自动下载失败，您可以手动下载V2Ray核心：")
	fmt.Println()
	fmt.Printf("1. 访问 V2Ray 官方发布页面:\n")
	fmt.Printf("   https://github.com/v2fly/v2ray-core/releases/latest\n")
	fmt.Println()
	fmt.Printf("2. 查找适合您系统的文件 (系统: %s-%s):\n", sysInfo.OS, sysInfo.Arch)

	// 根据系统提供具体的文件名建议
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			fmt.Printf("   建议下载: v2ray-macos-arm64.zip 或包含 'darwin' 和 'arm64' 的文件\n")
		} else {
			fmt.Printf("   建议下载: v2ray-macos-64.zip 或包含 'darwin' 和 '64' 的文件\n")
		}
	} else {
		fmt.Printf("   建议下载: v2ray-%s-%s.zip\n", sysInfo.OS, sysInfo.Arch)
	}

	fmt.Println()
	fmt.Printf("3. 下载后解压到: %s\n", d.InstallPath)
	fmt.Printf("4. 确保 v2ray 可执行文件位于: %s\n", filepath.Join(d.InstallPath, d.getV2rayExecutableName()))
	fmt.Println()
	fmt.Printf("5. 解压完成后，可以运行 '%s check-v2ray' 来验证安装\n", os.Args[0])
	fmt.Println()
}

// AutoDownloadV2Ray 自动下载V2Ray的便捷函数
func AutoDownloadV2Ray() error {
	downloader := NewV2RayDownloader()
	return downloader.DownloadAndInstall()
}
