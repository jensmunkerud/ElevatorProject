//go:build !windows

package processpair

import "syscall"

func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
