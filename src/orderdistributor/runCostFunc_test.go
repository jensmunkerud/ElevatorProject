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
	hallOrders := orders.CreateHallOrders()
	cabOrders := orders.CreateCabOrders()
	// Create a map for elevators
	elevators := make(map[string]*elevator.Elevator)
	elevators["bankID"] = elev1
	allCabOrders := make(map[string]*orders.CabOrders)
	allCabOrders["bankID"] = cabOrders
	elevatorsOnline := make(map[string]bool)
	// Create and initialize an elevator with dummy data
	elevatorsOnline["bankID"] = true

	input := make(chan any, 1)
	activeOrders := make(chan [][]bool, 1)

	config.MyID = "bankID"
	go runCostFunc(input, activeOrders)

	input <- CostFuncInput{
		AllCabOrders:     allCabOrders,
		MergedHallOrders: hallOrders,
		Elevators:        elevators,
	}

	ordersOut := <-activeOrders
	fmt.Printf("runCostFunc output: %+v\n", ordersOut)
	if ordersOut == nil {
		t.Fatalf("runCostFunc returned nil active orders")
	}
	if len(ordersOut) != config.NumFloors {
		t.Fatalf("expected %d floors, got %d", config.NumFloors, len(ordersOut))
	}
}
