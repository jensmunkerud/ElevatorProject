package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"fmt"
	"os/exec"
)

type CostFuncInput struct {
	AllCabOrders     map[string]*orders.CabOrders
	MergedHallOrders *orders.HallOrders
	Elevators        map[string]*elevator.Elevator
}

// splitCostFuncInput is a temporary adapter. Replace this when your final input structure is decided.
func splitCostFuncInput(raw any) (CostFuncInput, bool) {
	input, ok := raw.(CostFuncInput)
	return input, ok
}

func runCostFunc(
	input <-chan any,
	activeOrders chan<- [][]bool,
) {
	for raw := range input {
		parts, ok := splitCostFuncInput(raw)
		if !ok {
			activeOrders <- nil
			continue
		}

		// Formats data into JSON format
		jsonInput, err := ConvertToJson(config.MyID, parts.AllCabOrders, parts.MergedHallOrders, parts.Elevators)
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			activeOrders <- nil
			continue
		}

		// Executes hall_request_assigner command
		jsonOutput, err := executeCommand(jsonInput)
		if err != nil {
			fmt.Print("Error executing hall_request_assigner command")
			activeOrders <- nil
			continue
		}

		assignments, err := ConvertFromJson(jsonOutput)
		if err != nil {
			activeOrders <- nil
			continue
		}
		activeOrders <- assignments[config.MyID]
	}
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
