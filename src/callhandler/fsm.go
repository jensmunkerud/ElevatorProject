package callhandler

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"time"
)

func fsmInit(e *es.Elevator) {
	if controller.GetFloor() != -1 {
		if e.Behaviour() != es.Moving {
			e.UpdateInService(true)
			return
		}
	}
	mid := config.NumFloors / 2

	if e.CurrentFloor() < mid {
		controller.MoveElevatorUp()
		e.UpdateCurrentDirection(es.Up)
	} else {
		controller.MoveElevatorDown()
		e.UpdateCurrentDirection(es.Down)
	}
	e.UpdateBehaviour(es.Moving)
}

func fsmOnFloorArrival(
	e *es.Elevator,
	newFloor int,
	doorTimer *time.Timer,
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	e.UpdateCurrentFloor(newFloor)
	e.UpdateInService(true)
	// stopTimer(serviceTimer)

	switch e.Behaviour() {
	case es.Moving:
		// Necessary to restart the serviceTimer here, to also mark as out of service if door is obstructed.
		restartTimer(serviceTimer, config.ServiceTimeout)
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
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	fmt.Printf("\n\nfsmOnDoorTimeout()\n")
	if e.Obstruction() {
		// Keep door open
		restartTimer(doorTimer, config.DoorOpenDuration)
		e.UpdateInService(false)
		return
	}
	switch e.Behaviour() {
	case es.DoorOpen:
		newDirection, newBehaviour := requestsChooseDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)

		switch e.Behaviour() {
		case es.DoorOpen:
			e.UpdateInService(true)
			restartTimer(doorTimer, config.DoorOpenDuration)
			restartTimer(serviceTimer, config.ServiceTimeout)
			requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
		case es.Moving, es.Idle:
			restartTimer(serviceTimer, config.ServiceTimeout)
			e.UpdateInService(true)
			controller.SetDoorOpenLamp(false)
			if !e.StopPressed() && e.Behaviour() == es.Moving {
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
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	if e.Behaviour() != es.Idle ||
		e.CurrentFloor() != controller.GetFloor() {
		return
	}
	newDirection, newBehaviour := requestsChooseDirection(*e)
	e.UpdateCurrentDirection(newDirection)
	e.UpdateBehaviour(newBehaviour)
	e.UpdateInService(true)
	switch newBehaviour {
	case es.DoorOpen:
		controller.SetDoorOpenLamp(true)
		restartTimer(doorTimer, config.DoorOpenDuration)
		requestsClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
	case es.Moving:
		restartTimer(serviceTimer, config.ServiceTimeout)
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
