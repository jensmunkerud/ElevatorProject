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
	cabRequest   *orders.CabOrders
	hallRequests *orders.HallOrders
}

func CreateElevator(id string, currentFloor int, direction Direction, behaviour Behaviour) *Elevator {
	hallRequests := orders.CreateHallOrders(config.NumFloors)
	cabRequests := orders.CreateCabOrders(config.NumFloors)
	return &Elevator{
		id:           id,
		behaviour:    behaviour,
		floor:        currentFloor,
		direction:    direction,
		hallRequests: hallRequests,
		cabRequest:   cabRequests,
	}
}

func (e *Elevator) Id() string {
	return e.id
}

func (e *Elevator) Behaviour() Behaviour {
	// switch e.behaviour {
	// case Idle:
	// 	return "idle"
	// case Moving:
	// 	return "moving"
	// case DoorOpen:
	// 	return "doorOpen"
	// }
	// return "unknown"
	return e.behaviour
}

func (e *Elevator) UpdateBehaviour(behaviour Behaviour) {
	e.behaviour = behaviour
}

func (e *Elevator) CurrentFloor() int {
	return e.floor
}

func (e *Elevator) UpdateCurrentFloor(floor int) {
	e.floor = floor
}

func (e *Elevator) CurrentDirection() Direction {
	return e.direction
}

func (e *Elevator) UpdateCurrentDirection(direction Direction) {
	e.direction = direction
}

func (e *Elevator) CabRequests() *orders.CabOrders {
	return e.cabRequest
}

func (e *Elevator) HallRequests() *orders.HallOrders {
	return e.hallRequests
}
