package callhandler

import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/elevatorstruct"
	"fmt"
	"os/exec"
	"runtime"
)

func runCostFunc(hallRequests [config.NumFloors][2]bool, elevators map[string]*elevatorstruct.Elevator) (map[string]elevatorstruct.ElevatorButtons, error) {

	jsonString, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// EXECUTE COMMAND
	var cmd *exec.Cmd
	binaryPath := "./libs/project-resources-cost_fns/hall_request_assigner/hall_request_assigner"

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command(
			"cmd", "/c", "start", "cmd", "/k",
			binaryPath,
			"--input", jsonString,
		)
		fmt.Println("Running command on WINDOWS...")

	case "darwin":
		cmd = exec.Command(
			"osascript", "-e",
			`tell application "Terminal" to do script "`+
				binaryPath+` --input '`+jsonString+`'"`)
		fmt.Println("Running command on MAC...")

	case "linux":
		cmd = exec.Command(
			"gnome-terminal", "--",
			binaryPath,
			"--input", jsonString,
		)
		fmt.Println("Running command on LINUX...")

	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
		return map[string]elevatorstruct.ElevatorButtons{}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// READ TERMINAL OUTPUT
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// Convert bytes to string
	result := string(output)
	fmt.Println("Binary output as string:")
	fmt.Println(result)
	return ParseElevatorJson(result)
}
