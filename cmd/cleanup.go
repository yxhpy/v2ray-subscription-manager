package main

import (
	"fmt"
	"os"

	"github.com/yxhpy/v2ray-subscription-manager/internal/utils"
)

func main() {
	fmt.Printf("🧹 手动清理所有临时文件...\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// 显示当前工作目录
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("📁 当前目录: %s\n", wd)
	}

	// 执行清理
	utils.ForceCleanupAll()

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("✅ 清理完成！所有临时文件和状态文件已删除。\n")
	fmt.Printf("💡 提示：重新启动auto-proxy服务将会重新创建必要的文件。\n")
}
