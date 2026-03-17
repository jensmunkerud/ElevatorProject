package callhandler

import (
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
)

func requestsAbove(e es.Elevator) bool {
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
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Requests()[e.CurrentFloor()][btn] {
			return true
		}
	}
	return false
}

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

func requestsShouldStop(e es.Elevator) bool {
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

func requestsShouldClearImmediately(e es.Elevator, buttonFloor int, buttonType es.ButtonType) bool {
	return e.CurrentFloor() == buttonFloor &&
		((e.CurrentDirection() == es.Up && buttonType == es.HallUp) ||
			(e.CurrentDirection() == es.Down && buttonType == es.HallDown) ||
			e.CurrentDirection() == es.Stop ||
			buttonType == es.Cab)
}

func requestsClearAtCurrentFloor(
	e *es.Elevator,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	RequestUpdateCabOrder(e.CurrentFloor(), es.Cab, true, cabOrderUpdate)

	switch e.CurrentDirection() {
	case es.Up:
		e.UpdateRequest(e.CurrentFloor(), es.HallUp, false)
		RequestUpdateHallOrder(e.CurrentFloor(), es.HallUp, true, hallOrderUpdate)
		if !requestsAbove(*e) {
			e.UpdateRequest(e.CurrentFloor(), es.HallDown, false)
			RequestUpdateHallOrder(e.CurrentFloor(), es.HallDown, true, hallOrderUpdate)
		}

	case es.Down:
		e.UpdateRequest(e.CurrentFloor(), es.HallDown, false)
		RequestUpdateHallOrder(e.CurrentFloor(), es.HallDown, true, hallOrderUpdate)
		if !requestsBelow(*e) {
			e.UpdateRequest(e.CurrentFloor(), es.HallUp, false)
			RequestUpdateHallOrder(e.CurrentFloor(), es.HallUp, true, hallOrderUpdate)
		}

	default:
		e.UpdateRequest(e.CurrentFloor(), es.HallUp, false)
		e.UpdateRequest(e.CurrentFloor(), es.HallDown, false)
		RequestUpdateHallOrder(e.CurrentFloor(), es.HallUp, true, hallOrderUpdate)
		RequestUpdateHallOrder(e.CurrentFloor(), es.HallDown, true, hallOrderUpdate)
	}
}
