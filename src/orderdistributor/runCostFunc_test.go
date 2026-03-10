package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"fmt"
	"testing"
)

func TestCostFunc(t *testing.T) {
	// Initialize dummy hall requests (all false)
	elev1 := elevator.CreateElevator("bankID", 2, elevator.Down, elevator.Idle)
	hallOrders := orders.CreateHallOrders(config.NumFloors)
	cabOrders := orders.CreateCabOrders(config.NumFloors)
	// Create a map for elevators
	elevators := make(map[string]*elevator.Elevator)
	elevators["bankID"] = elev1
	allCabOrders := make(map[string]*orders.CabOrders)
	allCabOrders["bankID"] = cabOrders
	elevatorsOnline := make(map[string]bool)
	// Create and initialize an elevator with dummy data
	elevatorsOnline["bankID"] = true

	assigned, err := runCostFunc("bankID", allCabOrders, hallOrders, elevators)
	fmt.Printf("runCostFunc output: %+v, err: %v\n", assigned, err)
	if err != nil {
		t.Fatalf("runCostFunc failed: %v", err)
	}
}
