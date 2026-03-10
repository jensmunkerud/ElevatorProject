package orderdistributor

import (
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"fmt"
	"os/exec"
)

func runCostFunc(
	myId string,
	cabOrders map[string]*orders.CabOrders,
	hallOrders *orders.HallOrders,
	elevators map[string]*elevator.Elevator,
) (map[string]elevator.ElevatorButtons, error) {

	// Formats data into JSON format
	jsonInput, err := ConvertToJson(myId, cabOrders, hallOrders, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return map[string]elevator.ElevatorButtons{}, err
	}

	// Executes hall_request_assigner command
	jsonOutput, err := executeCommand(jsonInput)
	if err != nil {
		fmt.Print("Error executing hall_request_assigner command")
		return map[string]elevator.ElevatorButtons{}, err
	}

	// Formats JSON result into data
	return ConvertFromJson(jsonOutput)
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
