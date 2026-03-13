package callhandler

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"fmt"
	"time"
)

func setAllLights(es es.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ { // HARD CODED 2 BUTTONS, MAY BE BAD
			// es.Elevator_requestButtonLight(floor, btn, es.Requests()[floor][btn])
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.Requests()[floor][btn])
		}
	}
}

func fsmOnInitBetweenFloors(e *es.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.UpdateCurrentDirection(es.Down)
	e.UpdateBehaviour(es.Moving)
}

func fsmOnRequestButtonPress(e *es.Elevator, btnFloor int, btnType elevio.ButtonType) {

	switch e.Behaviour() {
	case es.DoorOpen:
		if requestsShouldClearImmediately(*e, btnFloor, btnType) {
			time.Sleep(config.DoorOpenDuration)
		} else {
			e.UpdateRequest(btnFloor, btnType, true)
		}

	case es.Moving:
		e.UpdateRequest(btnFloor, btnType, true)

	case es.Idle:
		e.UpdateRequest(btnFloor, btnType, true)
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)
		switch newBehaviour {
		case es.DoorOpen:
			elevio.SetDoorOpenLamp(true)
			time.Sleep(config.DoorOpenDuration)
			*e = requestsClearAtCurrentFloor(*e)

		case es.Moving:
			elevio.SetMotorDirection(elevio.MotorDirection(e.CurrentDirection()))

		case es.Idle:
			// nothing
		}
	}

	setAllLights(*e)

	fmt.Println("\nNew state!!")
}

func fsmOnFloorArrival(e *es.Elevator, newFloor int) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	e.UpdateCurrentFloor(newFloor)
	// elevatorFloorIndicator(e.floor) // whatafak

	switch e.Behaviour() {
	case es.Moving:
		if requestsShouldStop(*e) {
			elevio.SetMotorDirection(elevio.MotorDirection(e.CurrentDirection()))
			elevio.SetDoorOpenLamp(true)
			*e = requestsClearAtCurrentFloor(*e)
			time.Sleep(config.DoorOpenDuration)
			setAllLights(*e)
			e.UpdateBehaviour(es.DoorOpen)
		}
	default:
		// nothing
	}

	fmt.Println("\nNew state yea!")
}

func fsmOnDoorTimeout(e *es.Elevator) {
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")

	switch e.Behaviour() {
	case es.DoorOpen:
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)

		switch e.Behaviour() {
		case es.DoorOpen:
			time.Sleep(config.DoorOpenDuration)
			*e = requestsClearAtCurrentFloor(*e)
			setAllLights(*e)
		case es.Moving, es.Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(e.CurrentDirection()))
		}

	default:
		// nothing
	}

	fmt.Println("\nNew state again babbasjan!")
}
