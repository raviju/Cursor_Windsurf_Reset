//go:build !windows
// +build !windows

package cleaner

import (
	"os/exec"
	"strings"
)

func (e *Engine) isProcessRunning(processName string) bool {
	cmd := exec.Command("pgrep", "-i", processName)

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(processName))
}
