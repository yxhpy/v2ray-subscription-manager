package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
	CleanupHysteria2TempFiles()

	if cleanedCount > 0 {
		fmt.Printf("✅ 共清理了 %d 个临时文件\n", cleanedCount)
	} else {
		fmt.Printf("✅ 没有发现需要清理的临时文件\n")
	}
}

// CleanupHysteria2TempFiles 清理Hysteria2相关的临时文件
func CleanupHysteria2TempFiles() {
	hysteria2Patterns := []string{
		"./hysteria2/test_config_*.yaml",
		"./hysteria2/temp_*.yaml",
		"./hysteria2/config_*.yaml",
		"./hysteria2/test_proxy_*.yaml",
		"./hysteria2/proxy_server_*.yaml",
		"hysteria2/test_config_*.yaml", // 无./前缀的版本
		"hysteria2/temp_*.yaml",
		"hysteria2/config_*.yaml",
		"hysteria2/test_proxy_*.yaml",
		"hysteria2/proxy_server_*.yaml",
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

			if err := os.Remove(file); err == nil {
				fmt.Printf("    🗑️  已删除: %s\n", file)
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
		if err := os.Remove(file); err == nil {
			fmt.Printf("    🗑️  已删除: %s\n", file)
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
	CleanupTempFiles()
	CleanupAutoProxyFiles()
	fmt.Printf("✅ 强制清理完成\n")
}
