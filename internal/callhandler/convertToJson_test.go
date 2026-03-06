package callhandler

import (
	"encoding/json"
	"fmt"
	"testing"

	"elevatorproject/internal/config"
	es "elevatorproject/internal/elevatorstruct"
	ord "elevatorproject/internal/orders"
)

func TestConvertToJson_NewSignature_ProducesExpectedStructure(t *testing.T) {
	myID := ""

	myOrders := &es.Orders{
		CabOrders:  ord.CreateCabOrders(config.NumFloors),
		HallOrders: ord.CreateHallOrders(config.NumFloors),
	}
	*myOrders.HallOrders.Orders[0][0] = ord.OrderStateConfirmed
	*myOrders.HallOrders.Orders[1][1] = ord.OrderStateCompleted
	*myOrders.CabOrders.Orders[2] = ord.OrderStateConfirmed

	ordersMap := map[string]*es.Orders{
		myID: myOrders,
	}

	// A zero-value elevator is enough for this conversion test; id becomes "" -> state key "id_".
	elevators := map[string]*es.Elevator{
		"local": {},
	}

	isOnline := map[string]bool{
		myID: true,
	}

	jsonStr, err := ConvertToJson(myID, ordersMap, elevators, isOnline)
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
		t.Fatalf("expected HallRequests[1][1] to be true (Completed state)")
	}

	if len(message.States) != 1 {
		t.Fatalf("expected 1 elevator state, got %d", len(message.States))
	}

	state, ok := message.States["id_"]
	if !ok {
		t.Fatalf("expected state for key %q", "id_")
	}

	if len(state.CabRequests) != config.NumFloors {
		t.Fatalf("expected %d cab requests, got %d", config.NumFloors, len(state.CabRequests))
	}

	if !state.CabRequests[2] {
		t.Fatalf("expected CabRequests[2] to be true")
	}
}
