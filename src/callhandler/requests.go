package callhandler

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
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
		for btn := 0; btn < config.NumFloors; btn++ {
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
		return e.Requests()[e.CurrentFloor()][elevio.BT_HallDown] ||
			e.Requests()[e.CurrentFloor()][elevio.BT_Cab] ||
			!requestsBelow(e)
	case es.Up:
		return e.Requests()[e.CurrentFloor()][elevio.BT_HallUp] ||
			e.Requests()[e.CurrentFloor()][elevio.BT_Cab] ||
			!requestsAbove(e)
	default:
		return true
	}
}

func requestsShouldClearImmediately(e es.Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	return e.CurrentFloor() == btnFloor &&
		((e.CurrentDirection() == es.Up && btnType == elevio.BT_HallUp) ||
			(e.CurrentDirection() == es.Down && btnType == elevio.BT_HallDown) ||
			e.CurrentDirection() == es.Stop ||
			btnType == elevio.BT_Cab)
}

func requestsClearAtCurrentFloor(e es.Elevator) es.Elevator {
	e.UpdateRequest(e.CurrentFloor(), elevio.BT_Cab, false)

	switch e.CurrentDirection() {
	case es.Up:
		if !requestsAbove(e) && !e.Requests()[e.CurrentFloor()][elevio.BT_HallUp] {
			e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallDown, false)
		}
		e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallUp, false)

	case es.Down:
		if !requestsBelow(e) && !e.Requests()[e.CurrentFloor()][elevio.BT_HallUp] {
			e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallUp, false)
		}
		e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallDown, false)

	default:
		e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallUp, false)
		e.UpdateRequest(e.CurrentFloor(), elevio.BT_HallDown, false)
	}

	return e
}

