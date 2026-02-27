package elevatorstruct

import "elevatorproject/internal/config"

type ElevatorButtons struct {
	Buttons [config.NumFloors][2]bool //[up, down]
}

type Elevator struct {
	HallRequests ElevatorButtons
	id           string
	behaviour    string
	floor        int
	direction    string
	cabRequest   []bool
}

func (e *Elevator) Initialize(id string, currentFloor int, direction string) {
	e.id = id
	e.behaviour = "idle"
	e.floor = currentFloor
	e.direction = direction
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = make([]bool, config.NumFloors)
}

func (e *Elevator) CurrentElevatorId() string {
	return e.id
}

func (e *Elevator) Behaviour() string {
	return e.behaviour
}

func (e *Elevator) Floor() int {
	return e.floor
}

func (e *Elevator) Direction() string {
	return e.direction
}

func (e *Elevator) CabRequests() []bool {
	return e.cabRequest
}
