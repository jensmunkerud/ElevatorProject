package callhandler

import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/elevatorstruct"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func runCostFunc(hallRequests [config.NumFloors][2]bool, elevators map[string]*elevatorstruct.Elevator) (map[string]elevatorstruct.ElevatorButtons, error) {

	jsonString, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}
	command := "cd ../../libs/project-resources/cost_fns/hall_request_assigner; ./hall_request_assigner --input '" + jsonString + "'"
	result := executeCommand(command)
	fmt.Print(result)

	return map[string]elevatorstruct.ElevatorButtons{}, nil
}

func executeCommand(command string) string {
	// EXECUTE COMMAND
	var cmd *exec.Cmd
	// binaryPath := "../libs/project-resources-cost_fns/hall_request_assigner/hall_request_assigner"

	_, filename, _, _ := runtime.Caller(0)

	switch runtime.GOOS {

	case "windows":
		fmt.Println("Running command on WINDOWS...")
		// cmd = exec.Command(
		// 	"cmd", "/c", "start", "cmd", "/k",
		// 	binaryPath,
		// 	"--input", jsonString,
		// )

	case "darwin":
		fmt.Printf("Running command on MAC...\n")
		cmd = exec.Command("/bin/sh", "-c", command)
		cmd.Dir = filepath.Dir(filename)
		// READ TERMINAL OUTPUT
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error reading terminal:", output)
			return "error"
		}
		return string(output)

	case "linux":
		fmt.Println("Running command on LINUX...")
		// cmd = exec.Command(
		// 	"gnome-terminal", "--",
		// 	binaryPath,
		// 	"--input", jsonString,
		// )

	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
	}
	return "error"
}
