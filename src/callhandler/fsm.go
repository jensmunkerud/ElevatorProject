package callhandler

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"time"
)

func fsmOnInitBetweenFloors(e *es.Elevator) {
	controller.MoveElevatorDown()
	e.UpdateCurrentDirection(es.Down)
	e.UpdateBehaviour(es.Moving)
}

func fsmOnRequestButtonPress(
	e *es.Elevator,
	buttonFloor int,
	buttonType es.ButtonType,
	doorTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {

	switch e.Behaviour() {
	case es.DoorOpen:
		if requestsShouldClearImmediately(*e, buttonFloor, buttonType) {
			restartTimer(doorTimer, config.DoorOpenDuration)
			requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
		}

	// case es.Moving:
		// e.UpdateRequest(buttonFloor, buttonType, true)

	case es.Idle:
		// e.UpdateRequest(buttonFloor, buttonType, true)
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)
		switch newBehaviour {
		case es.DoorOpen:
			controller.SetDoorOpenLamp(true)
			restartTimer(doorTimer, config.DoorOpenDuration)
			requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)

		case es.Moving:
			if !e.StopPressed() {
				switch e.CurrentDirection() {
				case es.Up:
					controller.MoveElevatorUp()
				case es.Down:
					controller.MoveElevatorDown()
				}
			}

		case es.Idle:
			// nothing
		}
	}

	fmt.Println("\nNew state!!")
}

func fsmOnFloorArrival(
	e *es.Elevator,
	newFloor int,
	doorTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	e.UpdateCurrentFloor(newFloor)
	// elevatorFloorIndicator(e.floor) // whatafak

	switch e.Behaviour() {
	case es.Moving:
		if requestsShouldStop(*e) {
			controller.StopElevator()
			controller.SetDoorOpenLamp(true)
			requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
			restartTimer(doorTimer, config.DoorOpenDuration)
			e.UpdateBehaviour(es.DoorOpen)
		}
	default:
		// nothing
	}
}

func fsmOnDoorTimeout(
	e *es.Elevator,
	doorTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")
	if e.Obstruction() {
		// Keep door open
		restartTimer(doorTimer, config.DoorOpenDuration)
		return
	}
	switch e.Behaviour() {
	case es.DoorOpen:
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)

		switch e.Behaviour() {
		case es.DoorOpen:
			restartTimer(doorTimer, config.DoorOpenDuration)
			requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
		case es.Moving, es.Idle:
			controller.SetDoorOpenLamp(false)
			if !e.StopPressed() {
				switch e.CurrentDirection() {
				case es.Up:
					controller.MoveElevatorUp()
				case es.Down:
					controller.MoveElevatorDown()
				}
			}
		}

	default:
		// nothing
	}
}

func fsmOnNewOrders(
	e *es.Elevator,
	doorTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	if e.Behaviour() != es.Idle {
		return
	}
	newDir, newBeh := requestsChooseDirection(*e)
	e.UpdateCurrentDirection(newDir)
	e.UpdateBehaviour(newBeh)
	switch newBeh {
	case es.DoorOpen:
		controller.SetDoorOpenLamp(true)
		restartTimer(doorTimer, config.DoorOpenDuration)
		requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
	case es.Moving:
		if !e.StopPressed() {
			switch e.CurrentDirection() {
			case es.Up:
				controller.MoveElevatorUp()
			case es.Down:
				controller.MoveElevatorDown()
			}
		}
	}
}
