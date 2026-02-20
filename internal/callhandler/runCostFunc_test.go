package callhandler

import (
	es "elevatorproject/internal/elevatorstruct"
	"fmt"
	"testing"
)

func TestCostFunc(t *testing.T) {
	// Initialize dummy hall requests (all false)
	var hallRequests [4][2]bool

	// Create a map for elevators
	elevators := make(map[string]*es.Elevator)
	elev1 := &es.Elevator{}

	// Create and initialize an elevator with dummy data
	elev1.Initialize("bankID", 2, "down")
	elevators["bankID"] = elev1

	// Convert to JSON
	jsonString, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		fmt.Printf("Error converting to JSON: %v\n", err)
		return
	}

	// costFuncInput = jsonString
	fmt.Println("JSON String:", jsonString)
}
