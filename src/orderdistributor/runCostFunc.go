package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const costFuncInterval = 200 * time.Millisecond

// Run receives OrderDistributorMessages from the elevator server and runs the
// cost function at a throttled rate (costFuncInterval) to avoid spawning a
// subprocess on every network heartbeat.
func Run(
	input <-chan elevatorserver.OrderDistributorMessage,
	activeOrders chan<- [config.NumFloors][config.NumButtons]bool,
) {
	myID := config.MyID()
	fmt.Println("Starting orderdistributor loop")

	ticker := time.NewTicker(costFuncInterval)
	defer ticker.Stop()

	var latest *elevatorserver.OrderDistributorMessage

	for {
		select {
		case parts, ok := <-input:
			if !ok {
				return
			}
			latest = &parts

		case <-ticker.C:
			if latest == nil {
				continue
			}
			parts := *latest
			latest = nil

			allCabOrders, mergedHallOrders, elevators := parts.UnpackForOrderDistributor()

			// Remove non-servicable elevators from the cost function input
			for _, elevator := range elevators {
				id := elevator.Id()
				if !elevator.InService() {
					delete(elevators, id)
					delete(allCabOrders, id)
				}
			}
			if len(elevators) == 0 {
				fmt.Println("No elevators, skipping cost function")
				continue
			}
			jsonInput, err := ConvertToJson(config.MyID(), allCabOrders, mergedHallOrders, elevators)
			if err != nil {
				fmt.Printf("Error converting to JSON: %v\n", err)
				continue
			}

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
}

// hraDir returns the hall_request_assigner directory so it works whether
// the process is run from project root (e.g. go run main.go) or from src/orderdistributor.
func hraDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, rel := range []string{
		"libs/project-resources/cost_fns/hall_request_assigner",       // from project root (main.go)
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
	output, err := cmd.CombinedOutput()
	return string(output), err
}
