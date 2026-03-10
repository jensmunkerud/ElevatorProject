package orderdistributor

import (
	"encoding/json"
	"fmt"
	"testing"

	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	ord "elevatorproject/src/orders"
)

func TestConvertToJson_NewSignature_ProducesExpectedStructure(t *testing.T) {
	myID := ""

	// Create separate CabOrders and HallOrders
	cabOrders := map[string]*ord.CabOrders{
		myID: ord.CreateCabOrders(config.NumFloors),
	}
	*cabOrders[myID].Orders[2] = ord.ConfirmedOrderState

	hallOrders := ord.CreateHallOrders(config.NumFloors)
	*hallOrders.Orders[0][0] = ord.ConfirmedOrderState
	*hallOrders.Orders[1][1] = ord.ConfirmedOrderState

	// Create elevator map with pointer values
	elevators := map[string]*es.Elevator{
		myID: es.CreateElevator(myID, 2, es.Down, es.Moving),
	}

	jsonStr, err := ConvertToJson(myID, cabOrders, hallOrders, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}

	fmt.Println("Generated JSON:")
	fmt.Println(jsonStr)

	var message ElevatorMessage
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(message.HallRequests) != config.NumFloors {
		t.Fatalf("expected %d hall request rows, got %d", config.NumFloors, len(message.HallRequests))
	}

	if !message.HallRequests[0][0] {
		t.Fatalf("expected HallRequests[0][0] to be true")
	}

	if !message.HallRequests[1][1] {
		t.Fatalf("expected HallRequests[1][1] to be true")
	}

	if len(message.States) != 1 {
		t.Fatalf("expected 1 elevator state, got %d", len(message.States))
	}

	state, ok := message.States[myID]
	if !ok {
		t.Fatalf("expected state for key %q", myID)
	}

	if len(state.CabRequests) != config.NumFloors {
		t.Fatalf("expected %d cab requests, got %d", config.NumFloors, len(state.CabRequests))
	}

	if !state.CabRequests[2] {
		t.Fatalf("expected CabRequests[2] to be true")
	}
}
