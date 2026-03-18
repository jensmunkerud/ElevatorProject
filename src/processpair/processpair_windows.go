//go:build windows

package processpair

import "syscall"

const processQueryLimitedInformation = 0x1000

func isProcessAlive(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(h)
	return true
}
