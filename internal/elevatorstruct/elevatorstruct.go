package elevatorstruct

import (
	"elevatorproject/internal/config"
	"elevatorproject/internal/orders"
)

type Direction int

const (
	Stop Direction = 0
	Up   Direction = 1
	Down Direction = -1
)

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen
)

type ElevatorButtons struct {
	Buttons [config.NumFloors][2]bool //[up, down]
}

type Elevator struct {
	id           string
	behaviour    Behaviour
	floor        int
	direction    Direction
}

type Orders struct {
	id string
	CabOrders   *orders.CabOrders
	HallOrders  *orders.HallOrders
}

func createElevator(id string, currentFloor int, direction Direction, behaviour Behaviour) *Elevator {
	hallRequests := orders.CreateHallOrders(config.NumFloors)
	cabRequests := orders.CreateCabOrders(config.NumFloors)
	return &Elevator{
		id:           id,
		behaviour:    behaviour,
		floor:        currentFloor,
		direction:    direction,
		HallRequests: hallRequests,
		CabRequest:   cabRequests,
	}
}

func createOrders(id string) *Orders {
	hallRequests := orders.CreateHallOrders(config.NumFloors)
	cabRequests := orders.CreateCabOrders(config.NumFloors)
	return &Orders{
		id:        id,
		CabOrders: cabRequests,
		HallOrders: hallRequests,
	}
}

func (e *Elevator) CurrentElevatorId() string {
	return e.id
}

func (e *Elevator) Behaviour() string {
	switch e.behaviour {
	case Idle:
		return "idle"
	case Moving:
		return "moving"
	case DoorOpen:
		return "doorOpen"
}
	return "unknown"
}

func (e *Elevator) CurrentFloor() int {
	return e.floor
}

func (e *Elevator) CurrentDirection() Direction {
	return e.direction
}

func (o *Orders) CabRequests() *orders.CabOrders {
	return o.CabOrders
}

func (o *Orders) HallRequests() *orders.HallOrders {
	return o.HallOrders
}