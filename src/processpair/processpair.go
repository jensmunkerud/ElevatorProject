package processpair

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Run is called in backup mode (-processpair flag).
// It polls the master PID and returns when the master dies, so the caller can promote to master.
func Run(port int, masterPID int) {
	fmt.Printf("[Backup] Monitoring master (PID %d) for port %d...\n", masterPID, port)
	for {
		if !isProcessAlive(masterPID) {
			fmt.Printf("[Backup] Master (PID %d) for port %d died, promoting to master...\n", masterPID, port)
			time.Sleep(5 * time.Second)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// SpawnAndMonitorBackup spawns a backup process in a new terminal and monitors it.
// If the backup dies, a new one is spawned. Run as a goroutine from the master.
func SpawnAndMonitorBackup(port int) {
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("[Master] Could not resolve executable path: %v\n", err)
		return
	}
	fmt.Printf("[Master] Executable path: %s\n", executable)
	portStr := fmt.Sprintf("%d", port)
	myPID := fmt.Sprintf("%d", os.Getpid())

	for {
		fmt.Printf("[Master] Spawning backup for port %s\n", portStr)
		cmd := newBackupCmd(executable, portStr, myPID)
		if err := cmd.Run(); err != nil {
			fmt.Printf("[Master] Backup for port %s exited with error: %v\n", portStr, err)
		} else {
			fmt.Printf("[Master] Backup for port %s exited\n", portStr)
		}
		fmt.Printf("[Master] Respawning backup for port %s...\n", portStr)
		time.Sleep(5 * time.Second)
	}
}

func newBackupCmd(executable, portStr, masterPID string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "/wait",
			fmt.Sprintf("Backup %s", portStr),
			executable, "-port", portStr, "-processpair", "-masterpid", masterPID)
	case "linux":
		return exec.Command("gnome-terminal", "--wait", "--title",
			fmt.Sprintf("Backup %s", portStr), "--",
			executable, "-port", portStr, "-processpair", "-masterpid", masterPID)
	default:
		cmd := exec.Command(executable, "-port", portStr, "-processpair", "-masterpid", masterPID)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}
}
