package callhandler

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func runCostFunc() {

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", "go run main.go")
		fmt.Printf("Running command on WINDOWS...\n")
	case "darwin":
		wd, _ := os.Getwd()
		cmd = exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd `+wd+`; go run main.go"`)
		fmt.Printf("Running command on MAC...\n")
	case "linux":
		cmd = exec.Command("gnome-terminal", "--", "go", "run", "main.go")
		fmt.Printf("Running command on LINUX...\n")
	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
		return
	}
	cmd.Run()
}
