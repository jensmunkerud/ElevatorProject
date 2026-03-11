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
