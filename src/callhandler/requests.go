package callhandler

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
)

func requestsAbove(e elevator.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for floor := e.CurrentFloor() + 1; floor < config.NumFloors; floor++ {
		for orderType := range config.NumButtons {
			if e.Requests()[floor][orderType] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for floor := 0; floor < e.CurrentFloor(); floor++ {
		for orderType := range config.NumButtons {
			if e.Requests()[floor][orderType] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for orderType := range config.NumButtons {
		if e.Requests()[e.CurrentFloor()][orderType] {
			return true
		}
	}
	return false
}

// Find most suitable direction to move in, given current requests and direction.
func chooseNextDirection(e elevator.Elevator) (elevator.Direction, elevator.Behaviour) {
	switch e.CurrentDirection() {
	case elevator.Up:
		switch {
		case requestsAbove(e):
			return elevator.Up, elevator.Moving
		case requestsHere(e):
			return elevator.Down, elevator.DoorOpen
		case requestsBelow(e):
			return elevator.Down, elevator.Moving
		default:
			return elevator.Stop, elevator.Idle
		}
	case elevator.Down:
		switch {
		case requestsBelow(e):
			return elevator.Down, elevator.Moving
		case requestsHere(e):
			return elevator.Up, elevator.DoorOpen
		case requestsAbove(e):
			return elevator.Up, elevator.Moving
		default:
			return elevator.Stop, elevator.Idle
		}
	case elevator.Stop:
		switch {
		case requestsHere(e):
			return elevator.Stop, elevator.DoorOpen
		case requestsAbove(e):
			return elevator.Up, elevator.Moving
		case requestsBelow(e):
			return elevator.Down, elevator.Moving
		default:
			return elevator.Stop, elevator.Idle
		}
	default:
		return elevator.Stop, elevator.Idle
	}
}

// shouldStop returns true if the elevator should stop at the current floor to serve any orders.
func shouldStop(e elevator.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	switch e.CurrentDirection() {
	case elevator.Down:
		return e.Requests()[e.CurrentFloor()][elevator.HallDown] ||
			e.Requests()[e.CurrentFloor()][elevator.Cab] ||
			!requestsBelow(e)
	case elevator.Up:
		return e.Requests()[e.CurrentFloor()][elevator.HallUp] ||
			e.Requests()[e.CurrentFloor()][elevator.Cab] ||
			!requestsAbove(e)
	default:
		return true
	}
}

// orderIsAtCurrentStop checks if the incoming local order can immediately be cleared.
func orderIsAtCurrentStop(e elevator.Elevator, floor int, orderType elevator.OrderType) bool {
	if floor < 0 || floor >= config.NumFloors {
		return false
	}
	return e.CurrentFloor() == floor &&
		((e.CurrentDirection() == elevator.Up && orderType == elevator.HallUp) ||
			(e.CurrentDirection() == elevator.Down && orderType == elevator.HallDown) ||
			e.CurrentDirection() == elevator.Stop ||
			orderType == elevator.Cab)
}

func requestClearAtCurrentFloor(
	e *elevator.Elevator,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	requestUpdateOrder(e.CurrentFloor(), elevator.Cab, true, cabOrderUpdate, hallOrderUpdate)

	switch e.CurrentDirection() {
	case elevator.Up:
		if !requestsAbove(*e) && !e.Requests()[e.CurrentFloor()][elevator.HallUp] {
			requestUpdateOrder(e.CurrentFloor(), elevator.HallDown, true, cabOrderUpdate, hallOrderUpdate)
		}
		requestUpdateOrder(e.CurrentFloor(), elevator.HallUp, true, cabOrderUpdate, hallOrderUpdate)

	case elevator.Down:
		if !requestsBelow(*e) && !e.Requests()[e.CurrentFloor()][elevator.HallDown] {
			requestUpdateOrder(e.CurrentFloor(), elevator.HallUp, true, cabOrderUpdate, hallOrderUpdate)
		}
		requestUpdateOrder(e.CurrentFloor(), elevator.HallDown, true, cabOrderUpdate, hallOrderUpdate)

	default:
		requestUpdateOrder(e.CurrentFloor(), elevator.HallUp, true, cabOrderUpdate, hallOrderUpdate)
		requestUpdateOrder(e.CurrentFloor(), elevator.HallDown, true, cabOrderUpdate, hallOrderUpdate)
	}
}
