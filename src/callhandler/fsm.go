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
		for btn := 0; btn < 3; btn++ { // HARD CODED 2 BUTTONS, MAY BE BAD
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

func fsmOnRequestButtonPress(e *es.Elevator, btnFloor int, btnType elevio.ButtonType, doorTimer *time.Timer) {

	switch e.Behaviour() {
	case es.DoorOpen:
		if requestsShouldClearImmediately(*e, btnFloor, btnType) {
			startDoorTimer(doorTimer)
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
			startDoorTimer(doorTimer)
			*e = requestsClearAtCurrentFloor(*e)

		case es.Moving:
			if !e.StopPressed() {
				elevio.SetMotorDirection(elevio.MotorDirection(e.CurrentDirection()))
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
			elevio.SetMotorDirection(elevio.MotorDirection(es.Stop))
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
			if e.StopPressed() {
				elevio.SetMotorDirection(elevio.MotorDirection(e.CurrentDirection()))
			}
		}

	default:
		// nothing
	}

	fmt.Println("\nNew state again babbasjan!")
}

func startDoorTimer(t *time.Timer) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(config.DoorOpenDuration)
}
