package hallorders

import (
	orderstruct "elevatorproject/internal/orderStruct"
	"testing"
)

func TestCreateHallOrders_InitializesDimensionsAndUnknownState(t *testing.T) {
	numFloors := 4
	hallOrders := CreateHallOrders(numFloors)

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
			if got := hallOrders.HallOrderState(floor, direction); got != orderstruct.OrderStateUnknown {
				t.Logf("floor %d direction %d state: %v", floor, direction, got)
				t.Fatalf("floor %d direction %d: expected OrderStateUnknown, got %v", floor, direction, got)
			} else {
				t.Logf("floor %d direction %d state: %v", floor, direction, got)
			}
		}
	}
}

