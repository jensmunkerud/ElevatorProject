package elevator

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	"fmt"
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
	id        string
	behaviour Behaviour
	floor     int
	direction Direction
	requests  [config.NumFloors][config.NumButtons]bool
	activeOrders [][]bool // May be overlapping with requests. Double check with team.
}

func CreateElevator(id string, currentFloor int, direction Direction, behaviour Behaviour) *Elevator {
	return &Elevator{
		id:        id,
		behaviour: behaviour,
		floor:     currentFloor,
		direction: direction,
	}
}

// Sets corresponding light on floor, f -> floor, b -> 0 / 1 for down / up, on 1->on 0->off
func Elevator_requestButtonLight(f int, b int, on bool) {
	elevio.SetButtonLamp(2, f, on)
}

func (e *Elevator) Id() string {
	return e.id
}

func getID(elevator Elevator) string {
	return elevator.id
}

func (e *Elevator) Requests() [config.NumFloors][config.NumButtons]bool {
	return e.requests
}

func (e *Elevator) UpdateRequest(floor int, btn elevio.ButtonType, value bool) {
	e.requests[floor][btn] = value
}

func (e *Elevator) Behaviour() Behaviour {
	return e.behaviour
}

func (e *Elevator) BehaviourString() string {
	switch e.behaviour {
	case Idle:
		return "idle"
	case Moving:
		return "moving"
	case DoorOpen:
		return "doorOpen"
	}
	errorMsg := "unknown behaviour state: %v"
	return fmt.Sprintf(errorMsg, e.behaviour)
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

func (e *Elevator) DirectionString() string {
	switch e.direction {
	case Stop:
		return "stop"
	case Up:
		return "up"
	case Down:
		return "down"
	}
	return "unknown"
}

func (e *Elevator) UpdateCurrentDirection(direction Direction) {
	e.direction = direction
}

func (e *Elevator) UpdateActiveOrder(newActiveOrders [][]bool) {
	e.activeOrders = newActiveOrders
}
	
func (e *Elevator) Copy() Elevator {
    return *e
}