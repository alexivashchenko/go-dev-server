//go:build !windows
// +build !windows

package helpers

import (
	"os/exec"
	"syscall"
)

// setProcessGroupID configures a command to run in its own process group on Unix systems
func setProcessGroupID(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
