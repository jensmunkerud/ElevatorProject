package callhandler

import (
	"elevatorproject/internal/elevatorstruct"
	"fmt"
	"testing"
)

func TestCostFunc(t *testing.T) {
	// Initialize dummy hall requests (all false)
	elev1 := elevatorstruct.CreateElevator("bankID", 2, elevatorstruct.Down, elevatorstruct.Idle)
	orders := elevatorstruct.CreateOrders("bankID")
	// Create a map for elevators
	elevators := make(map[string]*elevatorstruct.Elevator)
	elevators["bankID"] = elev1
	ordersMap := make(map[string]*elevatorstruct.Orders)
	ordersMap["bankID"] = orders
	elevatorsOnline := make(map[string]bool)
	// Create and initialize an elevator with dummy data
	elevatorsOnline["bankID"] = true

	assigned, err := runCostFunc("bankID", ordersMap, elevators, elevatorsOnline)
	fmt.Printf("runCostFunc output: %+v, err: %v\n", assigned, err)
	if err != nil {
		t.Fatalf("runCostFunc failed: %v", err)
	}
}
