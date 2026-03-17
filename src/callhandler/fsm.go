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
			startDoorTimer(doorTimer)
		} else {
			e.UpdateRequest(buttonFloor, buttonType, true)
		}

	case es.Moving:
		e.UpdateRequest(buttonFloor, buttonType, true)

	case es.Idle:
		e.UpdateRequest(buttonFloor, buttonType, true)
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)
		switch newBehaviour {
		case es.DoorOpen:
			controller.SetDoorOpenLamp(true)
			startDoorTimer(doorTimer)
			*e = requestsClearAtCurrentFloor(*e, hallOrderUpdate, cabOrderUpdate)

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
			*e = requestsClearAtCurrentFloor(*e, hallOrderUpdate, cabOrderUpdate)
			startDoorTimer(doorTimer)
			e.UpdateBehaviour(es.DoorOpen)
		}
	default:
		// nothing
	}

	//fmt.Println("\nNew state yea!")
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
		startDoorTimer(doorTimer)
		return
	}
	switch e.Behaviour() {
	case es.DoorOpen:
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)

		switch e.Behaviour() {
		case es.DoorOpen:
			startDoorTimer(doorTimer)
			*e = requestsClearAtCurrentFloor(*e, hallOrderUpdate, cabOrderUpdate)
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

	//fmt.Println("New state again babbasjan!")
}

func startDoorTimer(t *time.Timer) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(config.DoorOpenDuration)
}
