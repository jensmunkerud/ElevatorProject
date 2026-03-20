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
	InService bool
}

// Message is the wire format for broadcasting worldview updates to the network.
type Message struct {
	SenderID       string
	HallOrders     [config.NumFloors][2]orders.OrderState
	AllCabOrders   map[string][config.NumFloors]orders.OrderState
	ElevatorStates map[string]ElevatorState
}

// messageFromWorldview builds a wire Message from the current worldview snapshots.
func messageFromWorldview(
	senderID string,
	hall *orders.HallOrders,
	allCab map[string]*orders.CabOrders,
	elevatorStates map[string]*elevator.Elevator,
) Message {
	msg := Message{
		SenderID:       senderID,
		AllCabOrders:   make(map[string][config.NumFloors]orders.OrderState, len(allCab)),
		ElevatorStates: make(map[string]ElevatorState, len(elevatorStates)),
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		for _, orderType := range elevator.HallOrderTypes {
			msg.HallOrders[floor][int(orderType)] = hall.GetOrderState(floor, orderType)
		}
	}
	for id, cab := range allCab {
		var arr [config.NumFloors]orders.OrderState
		for floor := 0; floor < config.NumFloors; floor++ {
			arr[floor] = cab.GetOrderState(floor)
		}
		msg.AllCabOrders[id] = arr
	}
	for id, elev := range elevatorStates {
		if elev == nil {
			continue
		}
		msg.ElevatorStates[id] = ElevatorState{
			Behaviour: elev.BehaviourToString(),
			Floor:     elev.CurrentFloor(),
			Direction: elev.DirectionString(),
			InService: elev.InService(),
		}
	}
	return msg
}
