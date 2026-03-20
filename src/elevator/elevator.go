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

type OrderType int

const (
	HallUp   OrderType = 0
	HallDown OrderType = 1
	Cab      OrderType = 2
)

type OrderEvent struct {
	Floor     int
	OrderType OrderType
}

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen
)

type Elevator struct {
	id          string
	behaviour   Behaviour
	floor       int
	direction   Direction
	requests    [config.NumFloors][config.NumButtons]bool
	obstruction bool
	stopPressed bool
	inService   bool
}

type ElevatorEvent struct {
	OrderEvent       chan OrderEvent
	FloorEvent       chan int
	ObstructionEvent chan bool
	StopEvent        chan bool
}

func CreateElevator(id string, currentFloor int, direction Direction, behaviour Behaviour) *Elevator {
	return &Elevator{
		id:        id,
		behaviour: behaviour,
		floor:     currentFloor,
		direction: direction,
		inService: false,
	}
}

func CreateElevatorEvent(
	orderEvent chan OrderEvent,
	floorEvent chan int,
	obstructionEvent chan bool,
	stopEvent chan bool,
) ElevatorEvent {
	return ElevatorEvent{
		OrderEvent:       orderEvent,
		FloorEvent:       floorEvent,
		ObstructionEvent: obstructionEvent,
		StopEvent:        stopEvent,
	}
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

func (e *Elevator) UpdateRequest(value [config.NumFloors][3]bool) {
	e.requests = value
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

func (e *Elevator) UpdateObstruction(obstructed bool) {
	e.obstruction = obstructed
}

func (e Elevator) Obstruction() bool {
	return e.obstruction
}

func (e *Elevator) UpdateStopPressed(stop bool) {
	e.stopPressed = stop
}

func (e Elevator) StopPressed() bool {
	return e.stopPressed
}

func (e *Elevator) UpdateInService(inService bool) {
	e.inService = inService
	if inService {
		fmt.Printf("Elevator is IN SERVICE!!\n")
	} else {
		fmt.Printf("Elevator is OUT OF SERVICE\n")
	}

}

func (e Elevator) InService() bool {
	return e.inService
}

func (e *Elevator) Copy() Elevator {
	return *e
}

func CreateOrderEvent(floor int, orderType OrderType) OrderEvent {
	return OrderEvent{
		Floor:     floor,
		OrderType: orderType,
	}
}