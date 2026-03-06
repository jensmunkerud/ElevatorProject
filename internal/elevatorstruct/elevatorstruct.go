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
	behaviour    behaviour
	floor        int
	direction    Direction
	CabRequest   *orders.CabOrders
	HallRequests *orders.HallOrders
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

func (e *Elevator) Id() string {
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

func (e *Elevator) CabRequests() *orders.CabOrders {
	return e.CabRequest
}

func (e *Elevator) HallRequests() *orders.HallOrders {
	return e.HallRequests
}