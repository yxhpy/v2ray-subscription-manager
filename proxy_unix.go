//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// setProcAttributes 设置进程属性（Unix平台）
func setProcAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
