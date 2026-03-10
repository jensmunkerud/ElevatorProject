package elevator

import (
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
	id           string
	behaviour    Behaviour
	floor        int
	direction    Direction
}

func CreateElevator(id string, currentFloor int, direction Direction, behaviour Behaviour) *Elevator {
	return &Elevator{
		id:           id,
		behaviour:    behaviour,
		floor:        currentFloor,
		direction:    direction,
	}
}

func (e *Elevator) Id() string {
	return e.id
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
