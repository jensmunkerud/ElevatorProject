package callhandler

import (
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
)

func requestsAbove(e es.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for f := e.CurrentFloor() + 1; f < config.NumFloors; f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests()[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e es.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for f := 0; f < e.CurrentFloor(); f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests()[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e es.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Requests()[e.CurrentFloor()][btn] {
			return true
		}
	}
	return false
}

// Find most suitable direction to move in, given current requests and direction.
func requestsChooseDirection(e es.Elevator) (es.Direction, es.Behaviour) {
	switch e.CurrentDirection() {
	case es.Up:
		switch {
		case requestsAbove(e):
			return es.Up, es.Moving
		case requestsHere(e):
			return es.Down, es.DoorOpen
		case requestsBelow(e):
			return es.Down, es.Moving
		default:
			return es.Stop, es.Idle
		}
	case es.Down:
		switch {
		case requestsBelow(e):
			return es.Down, es.Moving
		case requestsHere(e):
			return es.Up, es.DoorOpen
		case requestsAbove(e):
			return es.Up, es.Moving
		default:
			return es.Stop, es.Idle
		}
	case es.Stop:
		switch {
		case requestsHere(e):
			return es.Stop, es.DoorOpen
		case requestsAbove(e):
			return es.Up, es.Moving
		case requestsBelow(e):
			return es.Down, es.Moving
		default:
			return es.Stop, es.Idle
		}
	default:
		return es.Stop, es.Idle
	}
}

// shouldStop returns true if the elevator should stop at the current floor to serve any orders.
func shouldStop(e es.Elevator) bool {
	if e.CurrentFloor() < 0 || e.CurrentFloor() >= config.NumFloors {
		return false
	}
	switch e.CurrentDirection() {
	case es.Down:
		return e.Requests()[e.CurrentFloor()][es.HallDown] ||
			e.Requests()[e.CurrentFloor()][es.Cab] ||
			!requestsBelow(e)
	case es.Up:
		return e.Requests()[e.CurrentFloor()][es.HallUp] ||
			e.Requests()[e.CurrentFloor()][es.Cab] ||
			!requestsAbove(e)
	default:
		return true
	}
}

// orderIsAtCurrentStop checks if the incoming local order can immediately be cleared.
func orderIsAtCurrentStop(e es.Elevator, buttonFloor int, buttonType es.OrderType) bool {
	if buttonFloor < 0 || buttonFloor >= config.NumFloors {
		return false
	}
	return e.CurrentFloor() == buttonFloor &&
		((e.CurrentDirection() == es.Up && buttonType == es.HallUp) ||
			(e.CurrentDirection() == es.Down && buttonType == es.HallDown) ||
			e.CurrentDirection() == es.Stop ||
			buttonType == es.Cab)
}

func requestClearAtCurrentFloor(
	e *es.Elevator,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	RequestUpdateOrder(e.CurrentFloor(), es.Cab, true, cabOrderUpdate, hallOrderUpdate)

	switch e.CurrentDirection() {
	case es.Up:
		RequestUpdateOrder(e.CurrentFloor(), es.HallUp, true, cabOrderUpdate, hallOrderUpdate)
		if !requestsAbove(*e) {
			RequestUpdateOrder(e.CurrentFloor(), es.HallDown, true, cabOrderUpdate, hallOrderUpdate)
		}

	case es.Down:
		RequestUpdateOrder(e.CurrentFloor(), es.HallDown, true, cabOrderUpdate, hallOrderUpdate)
		if !requestsBelow(*e) {
			RequestUpdateOrder(e.CurrentFloor(), es.HallUp, true, cabOrderUpdate, hallOrderUpdate)
		}

	default:
		RequestUpdateOrder(e.CurrentFloor(), es.HallUp, true, cabOrderUpdate, hallOrderUpdate)
		RequestUpdateOrder(e.CurrentFloor(), es.HallDown, true, cabOrderUpdate, hallOrderUpdate)
	}
}
