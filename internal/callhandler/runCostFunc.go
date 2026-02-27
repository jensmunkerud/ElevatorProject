package callhandler

import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/elevatorstruct"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func runCostFunc(hallRequests [config.NumFloors][2]bool, elevators map[string]*elevatorstruct.Elevator) (map[string]elevatorstruct.ElevatorButtons, error) {

	// Formats data into JSON format
	jsonInput, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// Executes hall_request_assigner command
	command := "cd ../../libs/project-resources/cost_fns/hall_request_assigner; ./hall_request_assigner --input '" + jsonInput + "'"
	jsonOutput, err := executeCommand(command)
	if err != nil {
		fmt.Print("Error executing hall_request_assigner command")
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// Formats JSON result into data
	return ParseElevatorJson(jsonOutput)
}

func executeCommand(command string) (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("cmd", "/c", command)

	case "darwin":
		cmd = exec.Command("/bin/sh", "-c", command)

	case "linux":
		cmd = exec.Command("gnome-terminal", "--", command)

	default:
		return "", errors.New("Unsupported OS")
	}
	_, filename, _, _ := runtime.Caller(0)
	cmd.Dir = filepath.Dir(filename)
	// READ TERMINAL OUTPUT
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error reading terminal:", output)
		return "", errors.New("Error reading terminal")
	}
	fmt.Print(string(output))
	return string(output), nil
}
