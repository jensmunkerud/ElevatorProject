package orders

import (
	"testing"
)

func TestCreateCabOrders_InitializesLengthAndUnknownState(t *testing.T) {
	numFloors := 4
	cabOrders := CreateCabOrders()

	if cabOrders == nil {
		t.Fatal("CreateCabOrders returned nil")
	}

	if len(cabOrders.Orders) != numFloors {
		t.Fatalf("expected %d floors, got %d", numFloors, len(cabOrders.Orders))
	}

	for floor := 0; floor < numFloors; floor++ {
		if cabOrders.Orders[floor] == nil {
			t.Fatalf("expected non-nil order pointer at floor %d", floor)
		}
		if got := cabOrders.CabOrderState(floor); got != UnknownOrderState {
			t.Logf("floor %d state: %v", floor, got)
			t.Fatalf("floor %d: expected OrderStateUnknown, got %v", floor, got)
		} else {
			t.Logf("floor %d state: %v", floor, got)
		}
	}
}

func TestCabOrderState_ReadsUpdatedState(t *testing.T) {
	cabOrders := CreateCabOrders()

	updatedState := ConfirmedOrderState
	cabOrders.UpdateOrderState(1, updatedState)

	if got := cabOrders.CabOrderState(1); got != ConfirmedOrderState {
		t.Logf("updated floor 1 state: %v", got)
		t.Fatalf("expected OrderStateConfirmed, got %v", got)
	} else {
		t.Logf("updated floor 1 state: %v", got)
	}
}
