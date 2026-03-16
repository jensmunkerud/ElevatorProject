package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"os/exec"
)

// Receives OrderDistributorMessage from the elvator server, unpacks the message, converts to JSON, executes cost function, and sends active orders assigned to self
//
//	to call handler.
func Run(
	input <-chan elevatorserver.OrderDistributorMessage,
	activeOrders chan<- [][]bool,
) {
	for parts := range input {

		allCabOrders, mergedHallOrders, elevators := parts.UnpackForOrderDistributor()
		//Prevent running cost function if we have no elevators.
		if len(elevators) == 0 {
			continue
		}
		jsonInput, err := ConvertToJson(config.MyID(), allCabOrders, mergedHallOrders, elevators)
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			continue
		}

		// Executes hall_request_assigner command
		jsonOutput, err := executeCostFunction(jsonInput)
		if err != nil {
			continue
		}

		assignments, err := ConvertFromJson(jsonOutput)
		if err != nil {
			continue
		}
		activeOrders <- assignments[config.MyID()]
	}
}

func executeCostFunction(jsonInput string) (string, error) {
	cmd := exec.Command(
		"./hall_request_assigner",
		"--input",
		jsonInput,
	)
	cmd.Dir = "../../libs/project-resources/cost_fns/hall_request_assigner"
	output, err := cmd.CombinedOutput()
	return string(output), err
}
