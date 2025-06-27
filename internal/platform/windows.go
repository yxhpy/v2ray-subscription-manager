//go:build windows

package platform

import (
	"os/exec"
)

// SetProcAttributes 设置进程属性（Windows平台）
func SetProcAttributes(cmd *exec.Cmd) {
	// Windows平台不需要设置Setpgid
}
