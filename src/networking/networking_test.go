package networking

import (
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"testing"
)

func TestMessageFromOrders_RoundTrip(t *testing.T) {
	hall := orders.CreateHallOrders()
	hall.UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
	hall.UpdateOrderState(2, 1, orders.ConfirmedOrderState)

	cab := orders.CreateCabOrders()
	cab.UpdateOrderState(1, orders.ConfirmedOrderState)
	allCab := map[string]*orders.CabOrders{"elev-1": cab}

	localElev := *elevator.CreateElevator("elev-1", 2, elevator.Up, elevator.Moving)
	msg := messageFromOrders("elev-1", hall, allCab, localElev)

	if msg.SenderID != "elev-1" {
		t.Errorf("SenderID: got %q, want %q", msg.SenderID, "elev-1")
	}
	if got := orders.OrderState(msg.HallOrders[0][0]); got != orders.UnconfirmedOrderState {
		t.Errorf("HallOrders[0][0]: got %v, want %v", got, orders.UnconfirmedOrderState)
	}
	if got := orders.OrderState(msg.HallOrders[2][1]); got != orders.ConfirmedOrderState {
		t.Errorf("HallOrders[2][1]: got %v, want %v", got, orders.ConfirmedOrderState)
	}
	if got := orders.OrderState(msg.AllCabOrders["elev-1"][1]); got != orders.ConfirmedOrderState {
		t.Errorf("AllCabOrders[elev-1][1]: got %v, want %v", got, orders.ConfirmedOrderState)
	}
	state := msg.ElevatorStates["elev-1"]
	if state.Behaviour != "moving" {
		t.Errorf("ElevatorStates behaviour: got %q, want %q", state.Behaviour, "moving")
	}
	if state.Floor != 2 {
		t.Errorf("ElevatorStates floor: got %d, want %d", state.Floor, 2)
	}
	if state.Direction != "up" {
		t.Errorf("ElevatorStates direction: got %q, want %q", state.Direction, "up")
	}
}

func TestWorldviewFromMessage_RoundTrip(t *testing.T) {
	hall := orders.CreateHallOrders()
	hall.UpdateOrderState(1, 1, orders.CompletedOrderState)

	meCab := orders.CreateCabOrders()
	meCab.UpdateOrderState(3, orders.UnconfirmedOrderState)
	allCab := map[string]*orders.CabOrders{"elev-1": meCab}

	elevatorStates := map[string]*elevator.Elevator{
		"elev-1": elevator.CreateElevator("elev-1", 3, elevator.Down, elevator.Moving),
		"elev-2": elevator.CreateElevator("elev-2", 0, elevator.Stop, elevator.Idle),
	}

	msg := messageFromWorldview("elev-1", hall, allCab, elevatorStates)
	worldview := worldviewFromMessage(msg)
	gotCab, gotHall, gotElevators := worldview.UnpackForNetworking()

	if got := gotHall.GetOrderState(1, 1); got != orders.CompletedOrderState {
		t.Fatalf("hall round-trip mismatch: got %v", got)
	}
	if got := gotCab["elev-1"].GetOrderState(3); got != orders.UnconfirmedOrderState {
		t.Fatalf("cab round-trip mismatch: got %v", got)
	}
	if got := gotElevators["elev-1"].CurrentFloor(); got != 3 {
		t.Fatalf("elevator floor round-trip mismatch: got %d", got)
	}
	if got := gotElevators["elev-2"].Behaviour(); got != elevator.Idle {
		t.Fatalf("elevator behaviour round-trip mismatch: got %v", got)
	}
}
