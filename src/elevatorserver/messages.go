package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
)

// This file defines the message types and conversion functions used for communication between the
// elevator server and its consumers, including the call handler, order distributor, and networking components.

// HallOrderUpdate describes an incoming hall order event from another elevator.
type HallOrderUpdate struct {
	SenderID  string
	Floor     int
	Direction int
	State     orders.OrderState
}

// CabOrderUpdate describes an incoming cab order event from another elevator.
type CabOrderUpdate struct {
	SenderID string
	Floor    int
	State    orders.OrderState
}

// UnpackHallOrders unpacks a received HallOrders snapshot into individual
// HallOrderUpdate values, one per floor per direction, ready to send into hallUpdates.
func UnpackHallOrders(senderID string, hallOrders *orders.HallOrders) []HallOrderUpdate {
	if hallOrders == nil {
		return nil
	}
	updates := make([]HallOrderUpdate, 0, config.NumFloors*2)
	for floor := 0; floor < config.NumFloors; floor++ {
		for dir := 0; dir < 2; dir++ {
			updates = append(updates, HallOrderUpdate{
				SenderID:  senderID,
				Floor:     floor,
				Direction: dir,
				State:     hallOrders.GetOrderState(floor, dir),
			})
		}
	}
	return updates
}

// UnpackCabOrders unpacks a received allCabOrders map into individual
// CabOrderUpdate values, one per elevator per floor, ready to send into cabUpdates.
func UnpackCabOrders(allCabOrders map[string]*orders.CabOrders) []CabOrderUpdate {
	if allCabOrders == nil {
		return nil
	}
	updates := make([]CabOrderUpdate, 0, len(allCabOrders)*config.NumFloors)
	for elevID, cab := range allCabOrders {
		for floor := 0; floor < config.NumFloors; floor++ {
			updates = append(updates, CabOrderUpdate{
				SenderID: elevID,
				Floor:    floor,
				State:    cab.GetOrderState(floor),
			})
		}
	}
	return updates
}

type CallHandlerMessage struct {
	mergedHallOrders orders.HallOrders
	myCabOrders      orders.CabOrders
}

// UnpackForCallHandler returns pointer-based snapshots for call handler consumers.
func (m CallHandlerMessage) UnpackForCallHandler() (*orders.HallOrders, *orders.CabOrders) {
	hallOrders := m.mergedHallOrders.Copy()
	cabOrders := m.myCabOrders.Copy()
	return hallOrders, cabOrders
}

type OrderDistributorMessage struct {
	mergedHallOrders orders.HallOrders
	allCabOrders     map[string]orders.CabOrders
	elevatorState    map[string]elevator.Elevator
}

// UnpackForOrderDistributor returns pointer-based snapshots for order distributor consumers.
func (m OrderDistributorMessage) UnpackForOrderDistributor() (map[string]*orders.CabOrders, *orders.HallOrders, map[string]*elevator.Elevator) {
	allCabOrders := make(map[string]*orders.CabOrders, len(m.allCabOrders))
	for id, cab := range m.allCabOrders {
		allCabOrders[id] = cab.Copy()
	}

	hallOrders := m.mergedHallOrders.Copy()

	elevatorState := make(map[string]*elevator.Elevator, len(m.elevatorState))
	for id, elev := range m.elevatorState {
		elevCopy := elev
		elevatorState[id] = &elevCopy
	}

	return allCabOrders, hallOrders, elevatorState
}

type NetworkingDistributorMessage struct {
	// Consider changing order to keep consistency. Note any
	// errors that may arise.
	senderID         string
	allCabOrders     map[string]orders.CabOrders
	mergedHallOrders orders.HallOrders
	elevatorState    map[string]elevator.Elevator
}

func (m *NetworkingDistributorMessage) SenderID() string {
	return m.senderID
}

func NewNetworkingDistributorMessage(
	senderID string,
	allCabOrders map[string]*orders.CabOrders,
	hallOrders *orders.HallOrders,
	elevatorState map[string]*elevator.Elevator,
) NetworkingDistributorMessage {
	msg := NetworkingDistributorMessage{
		senderID:      senderID,
		allCabOrders:  make(map[string]orders.CabOrders, len(allCabOrders)),
		elevatorState: make(map[string]elevator.Elevator, len(elevatorState)),
	}
	if hallOrders != nil {
		msg.mergedHallOrders = *hallOrders.Copy()
	}
	for id, cab := range allCabOrders {
		if cab == nil {
			continue
		}
		msg.allCabOrders[id] = *cab.Copy()
	}
	for id, elev := range elevatorState {
		if elev == nil {
			continue
		}
		msg.elevatorState[id] = *elev
	}
	return msg
}

// UnpackForNetworking returns pointer-based snapshots for networking consumers.
func (m NetworkingDistributorMessage) UnpackForNetworking() (map[string]*orders.CabOrders, *orders.HallOrders, map[string]*elevator.Elevator) {
	allCabOrders := make(map[string]*orders.CabOrders, len(m.allCabOrders))
	for id, cab := range m.allCabOrders {
		allCabOrders[id] = cab.Copy()
	}

	hallOrders := m.mergedHallOrders.Copy()

	elevatorState := make(map[string]*elevator.Elevator, len(m.elevatorState))
	for id, elev := range m.elevatorState {
		elevCopy := elev
		elevatorState[id] = &elevCopy
	}

	return allCabOrders, hallOrders, elevatorState
}
