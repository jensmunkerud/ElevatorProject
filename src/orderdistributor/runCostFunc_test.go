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

func TestConvertToJsonOutput(t *testing.T) {
	elevID := "elev-1"

	cabOrders := map[string]*orders.CabOrders{
		elevID: orders.CreateCabOrders(),
	}
	cabOrders[elevID].UpdateOrderState(1, orders.ConfirmedOrderState)
	cabOrders[elevID].UpdateOrderState(3, orders.ConfirmedOrderState)

	hallOrders := orders.CreateHallOrders()
	hallOrders.UpdateOrderState(0, 0, orders.ConfirmedOrderState) // Floor 0, Up
	hallOrders.UpdateOrderState(2, 1, orders.ConfirmedOrderState) // Floor 2, Down

	elevators := map[string]*elevator.Elevator{
		elevID: elevator.CreateElevator(elevID, 1, elevator.Up, elevator.Moving),
	}

	jsonStr, err := ConvertToJson(elevID, cabOrders, hallOrders, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}

	fmt.Println("=== JSON input to cost function ===")
	fmt.Println(jsonStr)
}

func TestCostFunc(t *testing.T) {
	config.SetMyID()
	myID := config.MyID()

	elev1 := elevator.CreateElevator(myID, 2, elevator.Down, elevator.Idle)

	hallOrders := orders.CreateHallOrders()
	hallOrders.UpdateOrderState(0, 0, orders.ConfirmedOrderState)

	cabOrders := orders.CreateCabOrders()
	cabOrders.UpdateOrderState(3, orders.ConfirmedOrderState)

	elevators := make(map[string]*elevator.Elevator)
	elevators[myID] = elev1
	allCabOrders := make(map[string]*orders.CabOrders)
	allCabOrders[myID] = cabOrders

	input := make(chan es.OrderDistributorMessage, 1)
	activeOrders := make(chan [config.NumFloors][config.NumButtons]bool, 1)

	go Run(input, activeOrders)

	allCabOrdersValue := make(map[string]orders.CabOrders, len(allCabOrders))
	for id, cab := range allCabOrders {
		allCabOrdersValue[id] = *cab.Copy()
	}

	elevatorsValue := make(map[string]elevator.Elevator, len(elevators))
	for id, elev := range elevators {
		elevatorsValue[id] = *elev
	}

	jsonInput, err := ConvertToJson(myID, allCabOrders, hallOrders, elevators)
	if err != nil {
		t.Fatalf("ConvertToJson failed: %v", err)
	}
	fmt.Printf("=== Cost function input ===\n%s\n", jsonInput)

	input <- buildOrderDistributorMessage(*hallOrders.Copy(), allCabOrdersValue, elevatorsValue)

	ordersOut := <-activeOrders
	fmt.Printf("=== Cost function output ===\n")
	fmt.Printf("Active orders: %+v\n", ordersOut)

	// Verify floor 0 hall-up order was assigned
	if !ordersOut[0][0] {
		t.Errorf("expected hall-up order at floor 0 to be assigned")
	}
	// Verify cab order at floor 3 was assigned
	if !ordersOut[3][2] {
		t.Errorf("expected cab order at floor 3 to be assigned")
	}
}
