//go:build windows

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// SetProcAttributes 设置进程属性（Windows平台）
func SetProcAttributes(cmd *exec.Cmd) {
	// Windows平台设置CREATE_NEW_PROCESS_GROUP，便于进程管理
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// KillProcessByPort 通过端口杀死进程（Windows特有）
func KillProcessByPort(port int) error {
	// 使用netstat查找占用端口的进程
	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("执行netstat失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pidStr := fields[len(fields)-1]
				if pid, err := strconv.Atoi(pidStr); err == nil {
					return KillProcessByPID(pid)
				}
			}
		}
	}
	return nil
}

// KillProcessByPID 通过PID杀死进程（Windows特有）
func KillProcessByPID(pid int) error {
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	return cmd.Run()
}

// KillProcessByName 通过进程名杀死进程（Windows特有）
func KillProcessByName(name string) error {
	cmd := exec.Command("taskkill", "/F", "/IM", name)
	return cmd.Run()
}

// IsProcessRunning 检查进程是否在运行（Windows特有）
func IsProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), processName)
}
