package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	es "elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func buildOrderDistributorMessage(
	hallOrders orders.HallOrders,
	allCabOrders map[string]orders.CabOrders,
	elevators map[string]elevator.Elevator,
) es.OrderDistributorMessage {
	var msg es.OrderDistributorMessage
	v := reflect.ValueOf(&msg).Elem()

	setField := func(name string, value any) {
		f := v.FieldByName(name)
		reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
	}

	setField("mergedHallOrders", hallOrders)
	setField("allCabOrders", allCabOrders)
	setField("elevatorState", elevators)
	return msg
}

func TestCostFunc(t *testing.T) {
	// Initialize dummy hall requests (all false)
	elev1 := elevator.CreateElevator("bankID", 2, elevator.Down, elevator.Idle)
	hallOrders := orders.CreateHallOrders()
	cabOrders := orders.CreateCabOrders()
	// Create a map for elevators
	elevators := make(map[string]*elevator.Elevator)
	elevators["bankID"] = elev1
	allCabOrders := make(map[string]*orders.CabOrders)
	allCabOrders["bankID"] = cabOrders
	elevatorsOnline := make(map[string]bool)
	// Create and initialize an elevator with dummy data
	elevatorsOnline["bankID"] = true

	input := make(chan es.OrderDistributorMessage, 1)
	activeOrders := make(chan [][]bool, 1)

	config.MyID = "bankID"
	go RunCostFunc(input, activeOrders)

	allCabOrdersValue := make(map[string]orders.CabOrders, len(allCabOrders))
	for id, cab := range allCabOrders {
		allCabOrdersValue[id] = *cab.Copy()
	}

	elevatorsValue := make(map[string]elevator.Elevator, len(elevators))
	for id, elev := range elevators {
		elevatorsValue[id] = *elev
	}

	input <- buildOrderDistributorMessage(*hallOrders.Copy(), allCabOrdersValue, elevatorsValue)

	ordersOut := <-activeOrders
	fmt.Printf("runCostFunc output: %+v\n", ordersOut)
	if ordersOut == nil {
		t.Fatalf("runCostFunc returned nil active orders")
	}
	if len(ordersOut) != config.NumFloors {
		t.Fatalf("expected %d floors, got %d", config.NumFloors, len(ordersOut))
	}
}
