package callhandler

import (
	"encoding/json"
	"fmt"
	"testing"

	"elevatorproject/internal/config"
	es "elevatorproject/internal/elevatorstruct"
)

func TestConvertToJsonPrintsOutput(t *testing.T) {
	hallRequests := [config.NumFloors][2]bool{
		{true, false},
		{false, true},
		{false, false},
		{true, true},
	}

	elev1 := &es.Elevator{}
	elev1.Initialize("194234", 2, "up")

	elev2 := &es.Elevator{}
	elev2.Initialize("25435434", 0, "stop")

	elevators := map[string]*es.Elevator{
		"elev1": elev1,
		"elev2": elev2,
	}

	jsonStr, err := ConvertToJson(hallRequests, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}

	fmt.Println("\n=== ConvertToJson output ===")
	fmt.Println(jsonStr)
	fmt.Println("============================")

	var message ElevatorMessage
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(message.HallRequests) != config.NumFloors {
		t.Fatalf("Expected %d floors, got %d", config.NumFloors, len(message.HallRequests))
	}

	if len(message.States) != 2 {
		t.Fatalf("Expected 2 elevators, got %d", len(message.States))
	}
}
