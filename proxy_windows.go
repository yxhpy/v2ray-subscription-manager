//go:build windows

package main

import (
	"os/exec"
)

// setProcAttributes 设置进程属性（Windows平台）
func setProcAttributes(cmd *exec.Cmd) {
	// Windows平台不需要设置Setpgid
}
