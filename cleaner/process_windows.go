//go:build windows
// +build windows

package cleaner

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// isProcessRunning checks if a process is running on Windows
func (e *Engine) isProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(processName))
}
