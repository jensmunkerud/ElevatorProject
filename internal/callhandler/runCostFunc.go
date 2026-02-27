package callhandler

import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/elevatorstruct"
	"fmt"
	"os/exec"
)

func runCostFunc(hallRequests [config.NumFloors][2]bool, elevators map[string]*elevatorstruct.Elevator) (map[string]elevatorstruct.ElevatorButtons, error) {

	// Formats data into JSON format
	jsonInput, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// Executes hall_request_assigner command
	jsonOutput, err := executeCommand(jsonInput)
	if err != nil {
		fmt.Print("Error executing hall_request_assigner command")
		return map[string]elevatorstruct.ElevatorButtons{}, err
	}

	// Formats JSON result into data
	return ParseElevatorJson(jsonOutput)
}

func executeCommand(jsonInput string) (string, error) {
	cmd := exec.Command(
		"./hall_request_assigner",
		"--input",
		jsonInput,
	)

	cmd.Dir = "../../libs/project-resources/cost_fns/hall_request_assigner"

	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	return string(output), err
}
