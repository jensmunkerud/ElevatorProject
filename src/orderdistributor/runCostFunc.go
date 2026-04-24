package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"
)

func worldviewReadyForCostFunc(
	onlineNodes []string,
	allCabOrders map[string]*orders.CabOrders,
	elevators map[string]*elevator.Elevator,
) bool {
	if len(onlineNodes) == 0 {
		return false
	}
	for _, id := range onlineNodes {
		if _, ok := allCabOrders[id]; !ok {
			return false
		}
		elev, ok := elevators[id]
		if !ok || elev == nil {
			return false
		}
	}
	return true
}

// Run receives OrderDistributorMessages from the elevator server and runs the
// cost function at a throttled rate (costFuncInterval) to avoid spawning a
// subprocess on every network heartbeat.
func Run(
	input <-chan elevatorserver.OrderDistributorMessage,
	activeOrders chan<- [config.NumFloors][config.NumButtons]bool,
) {
	myID := config.MyID()
	fmt.Println("Starting orderdistributor loop")

	ticker := time.NewTicker(config.CostFuncInterval)
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
				fmt.Printf("No data for cost function yet, skipping\n")
				continue
			}
			parts := *latest
			latest = nil

			allCabOrders, mergedHallOrders, elevators, onlineNodes := parts.UnpackForOrderDistributor()

			if !worldviewReadyForCostFunc(onlineNodes, allCabOrders, elevators) {
				sort.Strings(onlineNodes)
				fmt.Printf("Skipping cost function: worldview not aligned yet. online=%v cab=%d elev=%d\n", onlineNodes, len(allCabOrders), len(elevators))
				continue
			}

			// Remove non-servicable elevators from the cost function input
			for _, elevator := range elevators {
				id := elevator.Id()
				isInService := elevator.InService()
				//fmt.Printf("State of elevator %v, %v\n", id, elevator)
				if !isInService {
					delete(elevators, id)
					delete(allCabOrders, id)
				}
			}
			if len(elevators) == 0 {
				fmt.Println("No elevators, skipping cost function")
				continue
			}
			jsonInput, err := ConvertToJson(allCabOrders, mergedHallOrders, elevators)
			if err != nil {
				fmt.Printf("Error converting to JSON: %v\n", err)
				continue
			}

			jsonOutput, err := executeCostFunction(jsonInput)
			if err != nil {
				fmt.Printf("Error executing cost function: %v\nOutput: %s\n", err, jsonOutput)
				continue
			}

			assignments, err := ConvertFromJson(jsonOutput)
			if err != nil {
				fmt.Printf("Error converting from JSON: %v\nOutput: %s\n", err, jsonOutput)
				continue
			}
			currentlyActive, ok := assignments[myID]
			if !ok {
				fmt.Printf("Cost function did not return an assignment for this elevator (ID %s)\n", myID)
				activeOrders <- [config.NumFloors][config.NumButtons]bool{}
				continue
			}
			activeOrders <- currentlyActive
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
		"--doorOpenDuration", strconv.FormatInt(config.DoorOpenDuration.Milliseconds(), 10),
		"--travelDuration", strconv.FormatInt(config.TravelDuration.Milliseconds(), 10),
	)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if len(output) == 0 {
		fmt.Printf("Cost func not activated")
		panic("Cost func not activated")
	}
	return string(output), err
}
