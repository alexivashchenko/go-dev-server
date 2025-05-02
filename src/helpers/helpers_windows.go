//go:build windows
// +build windows

package helpers

import (
	"os/exec"
)

// setProcessGroupID is a no-op on Windows
func setProcessGroupID(cmd *exec.Cmd) {
	// Windows doesn't support process groups in the same way
	// No action needed
}
