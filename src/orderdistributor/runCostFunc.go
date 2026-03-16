package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Receives OrderDistributorMessage from the elvator server, unpacks the message, converts to JSON, executes cost function, and sends active orders assigned to self
//
//	to call handler.
func Run(
	input <-chan elevatorserver.OrderDistributorMessage,
	activeOrders chan<- [config.NumFloors][config.NumButtons]bool,
) {
	myID := config.MyID()
	fmt.Println("Starting orderdistributor loop")
	for parts := range input {

		allCabOrders, mergedHallOrders, elevators := parts.UnpackForOrderDistributor()
		//Prevent running cost function if we have no elevators.
		if len(elevators) == 0 {
			fmt.Println("No elevators, skipping cost function")
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
		activeOrders <- assignments[myID]
	}
}

// hraDir returns the hall_request_assigner directory so it works whether
// the process is run from project root (e.g. go run main.go) or from src/orderdistributor.
func hraDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, rel := range []string{
		"libs/project-resources/cost_fns/hall_request_assigner",     // from project root (main.go)
		"../../libs/project-resources/cost_fns/hall_request_assigner", // from src/orderdistributor
	} {
		dir := filepath.Join(cwd, filepath.FromSlash(rel))
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("hall_request_assigner directory not found from %s", cwd)
}

func executeCostFunction(jsonInput string) (string, error) {
	fmt.Printf("Executing cost function with input: %s\n", jsonInput)
	dir, err := hraDir()
	if err != nil {
		return "", err
	}
	exe := "hall_request_assigner"
	if runtime.GOOS == "windows" {
		exe = "hall_request_assigner.exe"
	}
	cmd := exec.Command(
		filepath.Join(dir, exe),
		"--input", jsonInput,
		"--includeCab",
	)
	cmd.Dir = dir
	fmt.Printf("Command: %v\n", cmd)
	output, err := cmd.CombinedOutput()
	fmt.Printf("Cost function output: %s\n, %v", string(output), len(output))
	return string(output), err
}
