//go:build !windows

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// SetProcAttributes 设置进程属性（Unix平台）
func SetProcAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// KillProcessByPort 通过端口杀死进程（Unix平台）
func KillProcessByPort(port int) error {
	// 使用lsof查找占用端口的进程
	cmd := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		return nil // 端口未被占用或lsof不可用
	}

	pids := strings.Fields(string(output))
	for _, pidStr := range pids {
		if pid, err := strconv.Atoi(strings.TrimSpace(pidStr)); err == nil {
			KillProcessByPID(pid)
		}
	}
	return nil
}

// KillProcessByPID 通过PID杀死进程（Unix平台）
func KillProcessByPID(pid int) error {
	cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	return cmd.Run()
}

// KillProcessByName 通过进程名杀死进程（Unix平台）
func KillProcessByName(name string) error {
	cmd := exec.Command("pkill", "-f", name)
	return cmd.Run()
}

// IsProcessRunning 检查进程是否在运行（Unix平台）
func IsProcessRunning(processName string) bool {
	cmd := exec.Command("pgrep", "-f", processName)
	err := cmd.Run()
	return err == nil
}
