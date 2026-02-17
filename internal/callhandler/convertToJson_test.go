package callhandler

import (
	"encoding/json"
	"testing"

	es "Sanntid/internal/elevatorstruct"
)

func TestConvertToJson(t *testing.T) {
	// Setup test data
	hall := [4][2]bool{
		{true, false},
		{false, true},
		{false, false},
		{true, true},
	}

	elev1 := &es.Elevator{}
	elev1.Initialize(1, 2, "up")
	elev1.HallRequests = [4][2]bool{{false, false}, {true, false}, {false, false}, {false, true}}

	elev2 := &es.Elevator{}
	elev2.Initialize(2, 0, "stop")
	elev2.HallRequests = [4][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}

	elevators := map[string]*es.Elevator{
		"elev1": elev1,
		"elev2": elev2,
	}

	// Call function
	jsonStr, err := ConvertToJson(hall, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}

	// Verify JSON is valid
	var message ElevatorMessage
	err = json.Unmarshal([]byte(jsonStr), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Validate structure
	if len(message.HallRequests) != 4 {
		t.Errorf("Expected 4 floors, got %d", len(message.HallRequests))
	}

	if len(message.States) != 2 {
		t.Errorf("Expected 2 elevators, got %d", len(message.States))
	}

	// Validate elevator data
	if state, exists := message.States["elev1"]; exists {
		if state.Floor != 2 {
			t.Errorf("Expected floor 2, got %d", state.Floor)
		}
		if state.Behaviour != "idle" {
			t.Errorf("Expected behaviour 'idle', got %s", state.Behaviour)
		}
		if state.Direction != "up" {
			t.Errorf("Expected direction 'up', got %s", state.Direction)
		}
		if len(state.CabRequests) != 4 {
			t.Errorf("Expected 4 cab requests, got %d", len(state.CabRequests))
		}
	} else {
		t.Error("elev1 not found in states")
	}
}

func TestConvertToJsonPretty(t *testing.T) {
	hall := [4][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}
	elev := &es.Elevator{}
	elev.Initialize(1, 0, "stop")

	elevators := map[string]*es.Elevator{"elev1": elev}

	jsonStr, err := ConvertToJsonPretty(hall, elevators)
	if err != nil {
		t.Fatalf("ConvertToJsonPretty failed: %v", err)
	}

	// Verify it contains indentation (pretty format)
	if len(jsonStr) == 0 {
		t.Error("Expected non-empty JSON string")
	}

	// Verify it's valid JSON
	var message ElevatorMessage
	err = json.Unmarshal([]byte(jsonStr), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal pretty JSON: %v", err)
	}
}

func TestParseAssignerOutput(t *testing.T) {
	jsonOutput := `{
		"elev1": [[false, false], [true, false], [false, false], [false, true]],
		"elev2": [[false, false], [false, false], [false, false], [false, false]]
	}`

	output, err := ParseAssignerOutput(jsonOutput)
	if err != nil {
		t.Fatalf("ParseAssignerOutput failed: %v", err)
	}

	if len(output) != 2 {
		t.Errorf("Expected 2 elevators in output, got %d", len(output))
	}

	if requests, exists := output["elev1"]; exists {
		if len(requests) != 4 {
			t.Errorf("Expected 4 floors for elev1, got %d", len(requests))
		}
		if !requests[1][0] {
			t.Error("Expected requests[1][0] to be true")
		}
	} else {
		t.Error("elev1 not found in output")
	}
}

func TestUpdateHallRequests(t *testing.T) {
	// Create elevators
	elev1 := &es.Elevator{}
	elev1.Initialize(1, 0, "stop")
	elev1.HallRequests = [4][2]bool{{false, false}, {false, false}, {false, false}, {false, false}}

	elev2 := &es.Elevator{}
	elev2.Initialize(2, 0, "stop")

	elevators := map[string]*es.Elevator{
		"elev1": elev1,
		"elev2": elev2,
	}

	// Assigner output
	assignerOutput := AssignerOutput{
		"elev1": [][]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		"elev2": [][]bool{{true, false}, {false, false}, {false, false}, {false, false}},
	}

	// Update hall requests
	err := UpdateHallRequests(elevators, assignerOutput)
	if err != nil {
		t.Fatalf("UpdateHallRequests failed: %v", err)
	}

	// Verify elev1 updates
	if !elev1.HallRequests[1][0] {
		t.Error("Expected elev1 HallRequests[1][0] to be true")
	}
	if !elev1.HallRequests[3][1] {
		t.Error("Expected elev1 HallRequests[3][1] to be true")
	}

	// Verify elev2 updates
	if !elev2.HallRequests[0][0] {
		t.Error("Expected elev2 HallRequests[0][0] to be true")
	}
}

func TestUpdateHallRequestsInvalidLength(t *testing.T) {
	elev := &es.Elevator{}
	elev.Initialize(1, 0, "stop")

	elevators := map[string]*es.Elevator{"elev1": elev}

	// Invalid output: only 3 floors instead of 4
	assignerOutput := AssignerOutput{
		"elev1": [][]bool{{false, false}, {true, false}, {false, false}},
	}

	err := UpdateHallRequests(elevators, assignerOutput)
	if err == nil {
		t.Error("Expected error for invalid hall requests length")
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Create initial data
	hall := [4][2]bool{
		{true, false},
		{false, true},
		{false, false},
		{true, true},
	}

	elev1 := &es.Elevator{}
	elev1.Initialize(1, 2, "up")
	elev1.HallRequests = [4][2]bool{{false, false}, {true, false}, {false, false}, {false, true}}

	elevators := map[string]*es.Elevator{"elev1": elev1}

	// Convert to JSON and back
	jsonStr, err := ConvertToJson(hall, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}

	var message ElevatorMessage
	err = json.Unmarshal([]byte(jsonStr), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify round-trip integrity
	if message.HallRequests[0][0] != true || message.HallRequests[0][1] != false {
		t.Error("Hall requests not preserved in round-trip")
	}

	if message.States["elev1"].Floor != 2 {
		t.Error("Elevator floor not preserved in round-trip")
	}
}
