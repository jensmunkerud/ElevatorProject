package processpair

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Run acts as a backup watchdog for the elevator process.
// It continuously spawns the elevator with the given port and
// restarts it after a brief delay if it exits or crashes.
func Run(port int) {
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("[Backup] Could not resolve executable path: %v\n", err)
		return
	}

	portStr := fmt.Sprintf("%d", port)

	for {
		fmt.Printf("[Backup] Starting elevator on port %s\n", portStr)

		cmd := newElevatorCmd(executable, portStr)
		//cmd.Run() blocks untill the elevator process exits.
		err := cmd.Run()
		if err != nil {
			fmt.Printf("[Backup] Elevator on port %s exited with error: %v\n", portStr, err)
		} else {
			fmt.Printf("[Backup] Elevator on port %s exited\n", portStr)
		}

		time.Sleep(10 * time.Second)
		fmt.Printf("[Backup] Restarting elevator on port %s...\n", portStr)
	}
}

func newElevatorCmd(executable, portStr string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		// "start /wait" opens a new console window and blocks until it closes
		return exec.Command("cmd", "/c", "start", "/wait",
			fmt.Sprintf("Elevator %s", portStr),
			executable, "-port", portStr)
	case "linux":
		// gnome-terminal --wait opens a new terminal and blocks until it closes
		return exec.Command("gnome-terminal", "--wait", "--title",
			fmt.Sprintf("Elevator %s", portStr), "--",
			executable, "-port", portStr)
	default: 
		// macOS and others: run as direct child process
		cmd := exec.Command(executable, "-port", portStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd
	}
}
