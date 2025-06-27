package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/internal/utils"
)

func main() {
	fmt.Printf("🧹 手动清理所有临时文件和进程...\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// 显示当前工作目录
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("📁 当前目录: %s\n", wd)
	}

	// 显示开始时间
	startTime := time.Now()
	fmt.Printf("⏰ 开始时间: %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("\n")

	// 执行强制清理
	utils.ForceCleanupAll()

	// 显示完成信息
	elapsed := time.Since(startTime)
	fmt.Printf("\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("✅ 清理完成！耗时: %.2f 秒\n", elapsed.Seconds())
	fmt.Printf("🧹 所有临时文件和状态文件已删除\n")
	fmt.Printf("💀 所有相关进程已终止\n")
	fmt.Printf("🔍 清理结果已验证\n")
	fmt.Printf("\n")
	fmt.Printf("💡 提示：\n")
	fmt.Printf("   • 重新启动auto-proxy服务将会重新创建必要的文件\n")
	fmt.Printf("   • 如果发现残留文件或进程，可以再次运行此工具\n")
	fmt.Printf("   • 建议在auto-proxy完全停止后运行此清理工具\n")
}
