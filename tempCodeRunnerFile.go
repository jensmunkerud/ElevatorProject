import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/elevatorstruct"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Printf("Running command on MAC...\n")
	cmd = exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd `+wd+`; ls"`)
	// READ TERMINAL OUTPUT
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error reading terminal:", output)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}
	// Convert bytes to string
	result := string(output)
	fmt.Println("Binary output as string:")
	fmt.Println(result)
}