//go:build !windows

package platform

import (
	"os/exec"
	"syscall"
)

// SetProcAttributes 设置进程属性（Unix平台）
func SetProcAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
