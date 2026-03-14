package callhandler

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	controller "elevatorproject/src/controller"
	es "elevatorproject/src/elevator"
	"fmt"
	"time"
)

func setAllLights(es es.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := 0; button < 3; button++ { // HARD CODED 2 BUTTONS, MAY BE BAD
			// es.Elevator_requestButtonLight(floor, btn, es.Requests()[floor][btn])
			controller.SetButtonLamp(es.ButtonType(button), floor, es.Requests()[floor][button])
		}
	}
}

func fsmOnInitBetweenFloors(e *es.Elevator) {
	controller.MoveElevatorDown()
	e.UpdateCurrentDirection(es.Down)
	e.UpdateBehaviour(es.Moving)
}

func fsmOnRequestButtonPress(e *es.Elevator, buttonFloor int, buttonType es.ButtonType, doorTimer *time.Timer) {

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
			elevio.SetDoorOpenLamp(true)
			startDoorTimer(doorTimer)
			*e = requestsClearAtCurrentFloor(*e)

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

	setAllLights(*e)

	fmt.Println("\nNew state!!")
}

func fsmOnFloorArrival(e *es.Elevator, newFloor int, doorTimer *time.Timer) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	e.UpdateCurrentFloor(newFloor)
	// elevatorFloorIndicator(e.floor) // whatafak

	switch e.Behaviour() {
	case es.Moving:
		if requestsShouldStop(*e) {
			controller.StopElevator()
			elevio.SetDoorOpenLamp(true)
			*e = requestsClearAtCurrentFloor(*e)
			startDoorTimer(doorTimer)
			setAllLights(*e)
			e.UpdateBehaviour(es.DoorOpen)
		}
	default:
		// nothing
	}

	fmt.Println("\nNew state yea!")
}

func fsmOnDoorTimeout(e *es.Elevator, doorTimer *time.Timer) {
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
			*e = requestsClearAtCurrentFloor(*e)
			setAllLights(*e)
		case es.Moving, es.Idle:
			elevio.SetDoorOpenLamp(false)
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

	fmt.Println("New state again babbasjan!")
}

func startDoorTimer(t *time.Timer) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(config.DoorOpenDuration)
}
