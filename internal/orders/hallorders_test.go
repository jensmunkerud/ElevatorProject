package orders

import (
	"fmt"
	"testing"
)

func TestCreateHallOrders_InitializesDimensionsAndUnknownState(t *testing.T) {
	numFloors := 4
	hallOrders := CreateHallOrders()

	if hallOrders == nil {
		t.Fatal("CreateHallOrders returned nil")
	}

	if len(hallOrders.Orders) != numFloors {
		t.Fatalf("expected %d floors, got %d", numFloors, len(hallOrders.Orders))
	}

	for floor := 0; floor < numFloors; floor++ {
		for direction := 0; direction < 2; direction++ {
			if hallOrders.Orders[floor][direction] == nil {
				t.Fatalf("expected non-nil pointer at floor %d direction %d", floor, direction)
			}
			if got := hallOrders.HallOrderState(floor, direction); got != UnknownOrderState {
				t.Logf("floor %d direction %d state: %v", floor, direction, got)
				t.Fatalf("floor %d direction %d: expected OrderStateUnknown, got %v", floor, direction, got)
			} else {
				t.Logf("floor %d direction %d state: %v", floor, direction, got)
			}
		}
	}
}

func TestSimplifyHallOrders(t *testing.T) {
	hallOrders := CreateHallOrders()

	// Set some orders to different states
	hallOrders.UpdateOrderState(0, 0, ConfirmedOrderState) // Floor 0, Up - should be true (active)
	hallOrders.UpdateOrderState(0, 1, CompletedOrderState)   // Floor 0, Down - should be true (transition state)
	hallOrders.UpdateOrderState(2, 0, RemovedOrderState)     // Floor 2, Up - should be false
	hallOrders.UpdateOrderState(3, 1, UnconfirmedOrderState) // Floor 3, Down - should be false

	// Print the original states
	fmt.Println("Original Hall Order States:")
	stateNames := map[OrderState]string{
		UnknownOrderState:     "Unknown",
		RemovedOrderState:     "NoOrder",
		UnconfirmedOrderState: "Unconfirmed",
		ConfirmedOrderState:   "Confirmed",
		CompletedOrderState:   "Completed",
	}
	for floor := 0; floor < len(hallOrders.Orders); floor++ {
		upState := hallOrders.HallOrderState(floor, 0)
		downState := hallOrders.HallOrderState(floor, 1)
		fmt.Printf("Floor %d: [Up=%s, Down=%s]\n", floor, stateNames[upState], stateNames[downState])
	}

	// Call Simplify method
	simplified := hallOrders.Simplify()

	// Print the bool array
	fmt.Println("\nSimplified Hall Orders:")
	for floor := 0; floor < len(simplified); floor++ {
		fmt.Printf("Floor %d: [Up=%t, Down=%t]\n", floor, simplified[floor][0], simplified[floor][1])
	}

	// Test assertions
	tests := []struct {
		floor     int
		direction int
		expected  bool
		state     string
	}{
		{0, 0, true, "OrderStateConfirmed"},
		{0, 1, true, "OrderStateCompleted"},
		{2, 0, false, "OrderStateNoOrder"},
		{3, 1, false, "OrderStateUnconfirmed"},
	}

	for _, tt := range tests {
		if simplified[tt.floor][tt.direction] != tt.expected {
			t.Errorf("Floor %d Direction %d (%s): expected %t, got %t", tt.floor, tt.direction, tt.state, tt.expected, simplified[tt.floor][tt.direction])
		}
	}

	t.Logf("Simplified Hall Orders:")
	for floor := 0; floor < len(simplified); floor++ {
		t.Logf("Floor %d: [Up=%t, Down=%t]", floor, simplified[floor][0], simplified[floor][1])
	}
}
