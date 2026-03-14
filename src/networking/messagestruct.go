package networking

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
)

// ElevatorState is the wire representation of a single elevator's physical state.
type ElevatorState struct {
	Behaviour string
	Floor     int
	Direction string
}

// ElevatorStateUpdate pairs a sender ID with the received elevator state,
// for use in the elevatorStatesOut channel.
type ElevatorStateUpdate struct {
	SenderID string
	State    ElevatorState
}

// Message is the wire format for inter-elevator broadcasts.
// HallOrders[floor][dir] and AllCabOrders[elevID][floor] store OrderState values as ints
// so that JSON encoding works without requiring exported fields on the domain types.
// ElevatorStates carries the physical state of every known elevator, keyed by ID.
type Message struct {
	SenderID       string
	HallOrders     [config.NumFloors][2]int
	AllCabOrders   map[string][config.NumFloors]int
	ElevatorStates map[string]ElevatorState
}

// messageFromOrders builds a wire Message from the current hall, cab, and elevator state snapshots.
func messageFromOrders(senderID string, hall *orders.HallOrders, allCab map[string]*orders.CabOrders, elev elevator.Elevator) Message {
	return messageFromWorldview(senderID, hall, allCab, map[string]*elevator.Elevator{senderID: &elev})
}

func messageFromWorldview(
	senderID string,
	hall *orders.HallOrders,
	allCab map[string]*orders.CabOrders,
	elevatorStates map[string]*elevator.Elevator,
) Message {
	msg := Message{
		SenderID:       senderID,
		AllCabOrders:   make(map[string][config.NumFloors]int, len(allCab)),
		ElevatorStates: make(map[string]ElevatorState, len(elevatorStates)),
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		for dir := 0; dir < 2; dir++ {
			msg.HallOrders[floor][dir] = int(hall.GetOrderState(floor, dir))
		}
	}
	for id, cab := range allCab {
		var arr [config.NumFloors]int
		for floor := 0; floor < config.NumFloors; floor++ {
			arr[floor] = int(cab.GetOrderState(floor))
		}
		msg.AllCabOrders[id] = arr
	}
	for id, elev := range elevatorStates {
		if elev == nil {
			continue
		}
		msg.ElevatorStates[id] = ElevatorState{
			Behaviour: elev.BehaviourString(),
			Floor:     elev.CurrentFloor(),
			Direction: elev.DirectionString(),
		}
	}
	return msg
}
