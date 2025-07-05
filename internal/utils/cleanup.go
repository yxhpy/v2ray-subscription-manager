package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CleanupTempFiles 清理临时文件的通用函数
func CleanupTempFiles() {
	fmt.Printf("🧹 开始清理临时文件...\n")

	// 所有可能的临时文件模式
	patterns := []string{
		// V2Ray相关临时文件
		"temp_v2ray_config_*.json",
		"test_v2ray_config_*.json",
		"proxy_server_v2ray_*.json",
		"temp_config_*.json",
		"test_proxy_*.json",

		// 智能代理测试文件
		"temp_config_test_*.json",
		"intelligent_proxy_config_*.json",
		"intelligent_proxy_test_*.json",
		"intelligent_proxy_*.json",

		// 通用临时文件
		"*.tmp",
		"*.temp",

		// Auto-proxy相关文件
		"auto_proxy_best_node.json",
		"auto_proxy_state.json",
		"valid_nodes.json",
		"mvp_best_node.json",
		"proxy_state.json",
	}

	cleanedCount := 0
	for _, pattern := range patterns {
		cleanedCount += cleanupPattern(pattern)
	}

	// 清理hysteria2目录下的临时文件
	CleanupHysteria2TempFiles()

	if cleanedCount > 0 {
		fmt.Printf("✅ 共清理了 %d 个临时文件\n", cleanedCount)
	} else {
		fmt.Printf("✅ 没有发现需要清理的临时文件\n")
	}
}

// cleanupPattern 清理指定模式的文件，支持重试
func cleanupPattern(pattern string) int {
	cleanedCount := 0
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0
	}

	for _, file := range matches {
		if cleanupFileWithRetry(file) {
			cleanedCount++
		}
	}

	return cleanedCount
}

// cleanupFileWithRetry 删除文件，支持重试机制
func cleanupFileWithRetry(file string) bool {
	// 首先检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false // 文件不存在，无需删除
	}

	maxRetries := 3
	retryDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		if err := os.Remove(file); err == nil {
			fmt.Printf("    🗑️  已删除: %s\n", file)
			return true
		} else if os.IsNotExist(err) {
			// 文件在删除过程中被其他进程删除了
			return false
		} else if i == maxRetries-1 {
			fmt.Printf("    ❌ 删除失败: %s - %v\n", file, err)
			return false
		} else {
			fmt.Printf("    🔄 重试删除: %s (第%d次)\n", file, i+1)
			time.Sleep(retryDelay)
		}
	}

	return false
}

// CleanupHysteria2TempFiles 清理Hysteria2相关的临时文件
func CleanupHysteria2TempFiles() {
	hysteria2Patterns := []string{
		"./hysteria2/test_config_*.yaml",
		"./hysteria2/temp_*.yaml",
		"./hysteria2/config_*.yaml",
		"./hysteria2/test_proxy_*.yaml",
		"./hysteria2/proxy_server_*.yaml",
		"./hysteria2/intelligent_proxy_*.yaml",
		"hysteria2/test_config_*.yaml", // 无./前缀的版本
		"hysteria2/temp_*.yaml",
		"hysteria2/config_*.yaml",
		"hysteria2/test_proxy_*.yaml",
		"hysteria2/proxy_server_*.yaml",
		"hysteria2/intelligent_proxy_*.yaml",
	}

	cleanedCount := 0
	for _, pattern := range hysteria2Patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range matches {
			// 跳过主配置文件
			if file == "./hysteria2/config.yaml" || file == "hysteria2/config.yaml" {
				continue
			}

			if cleanupFileWithRetry(file) {
				cleanedCount++
			}
		}
	}
}

// CleanupAutoProxyFiles 清理Auto-proxy相关的文件
func CleanupAutoProxyFiles() {
	fmt.Printf("🧹 清理Auto-proxy相关文件...\n")

	files := []string{
		"auto_proxy_best_node.json",
		"auto_proxy_state.json",
		"valid_nodes.json",
		"mvp_best_node.json",
		"proxy_state.json",
	}

	cleanedCount := 0
	for _, file := range files {
		if cleanupFileWithRetry(file) {
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		fmt.Printf("✅ 清理了 %d 个Auto-proxy文件\n", cleanedCount)
	}
}

// ForceCleanupAll 强制清理所有临时文件和状态文件
func ForceCleanupAll() {
	fmt.Printf("🧹 执行强制清理...\n")

	// 第一步：终止相关进程
	KillRelatedProcesses()

	// 第二步：等待进程完全终止
	time.Sleep(2 * time.Second)

	// 第三步：清理临时文件
	CleanupTempFiles()

	// 第四步：清理Auto-proxy文件
	CleanupAutoProxyFiles()

	// 第五步：验证清理结果
	VerifyCleanup()

	fmt.Printf("✅ 强制清理完成\n")
}

// KillRelatedProcesses 终止相关进程
func KillRelatedProcesses() {
	fmt.Printf("💀 终止相关进程...\n")

	processNames := []string{"v2ray", "xray", "hysteria2", "hysteria"}

	for _, processName := range processNames {
		// 首先尝试温和终止
		cmd := exec.Command("pkill", "-TERM", "-f", processName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("    📤 发送终止信号给 %s 进程\n", processName)
		}
	}

	// 等待一段时间让进程优雅退出
	time.Sleep(3 * time.Second)

	// 强制终止仍在运行的进程
	for _, processName := range processNames {
		cmd := exec.Command("pkill", "-KILL", "-f", processName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("    💀 强制终止 %s 进程\n", processName)
		}
	}
}

// VerifyCleanup 验证清理结果
func VerifyCleanup() {
	fmt.Printf("🔍 验证清理结果...\n")

	// 检查关键文件是否仍存在
	criticalFiles := []string{
		"auto_proxy_best_node.json",
		"auto_proxy_state.json",
		"valid_nodes.json",
		"mvp_best_node.json",
	}

	remainingFiles := 0
	for _, file := range criticalFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("    ⚠️ 文件仍存在: %s\n", file)
			remainingFiles++
		}
	}

	// 检查是否有进程仍在运行
	processNames := []string{"v2ray", "xray", "hysteria2"}
	runningProcesses := 0
	for _, processName := range processNames {
		if isProcessRunning(processName) {
			fmt.Printf("    ⚠️ 进程仍在运行: %s\n", processName)
			runningProcesses++
		}
	}

	if remainingFiles == 0 && runningProcesses == 0 {
		fmt.Printf("    ✅ 清理验证通过\n")
	} else {
		fmt.Printf("    ⚠️ 发现 %d 个残留文件，%d 个残留进程\n", remainingFiles, runningProcesses)
	}
}

// isProcessRunning 检查进程是否仍在运行
func isProcessRunning(processName string) bool {
	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.Output()
	return err == nil && len(output) > 0
}
