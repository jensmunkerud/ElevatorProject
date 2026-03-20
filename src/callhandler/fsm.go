package callhandler

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"time"
)

func initializeElevatorToValidFloor(e *es.Elevator) {
	if config.NumFloors < 1 {
		fmt.Println("Error: NumFloors must be at least 1")
		return
	}
	floor := controller.GetFloor()
	if floor != -1 && floor >= 0 && floor < config.NumFloors {
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

func elevatorArrivedAtFloor(
	e *es.Elevator,
	newFloor int,
	doorTimer *time.Timer,
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	if newFloor < 0 || newFloor >= config.NumFloors {
		fmt.Printf("Error: Invalid floor %d\n", newFloor)
		return
	}
	fmt.Printf("\n\nfsmOnFloorArrival(%d)\n", newFloor)

	e.UpdateCurrentFloor(newFloor)
	e.UpdateInService(true)
	controller.SetFloorIndicator(newFloor)

	switch e.Behaviour() {
	case es.Moving:
		// Necessary to restart the serviceTimer here, to also mark as out of service if door is obstructed.
		restartTimer(serviceTimer, config.ServiceTimeout)
		if shouldStop(*e) {
			controller.StopElevator()
			controller.SetDoorOpenLamp(true)
			requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
			restartTimer(doorTimer, config.DoorOpenDuration)
			e.UpdateBehaviour(es.DoorOpen)
		}
	default:
		// nothing
	}
}

func elevatorOnDoorTimeout(
	e *es.Elevator,
	doorTimer *time.Timer,
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	if e.Obstruction() {
		// Keep door open
		restartTimer(doorTimer, config.DoorOpenDuration)
		restartTimer(serviceTimer, config.ServiceTimeout)
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
			requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
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

func elevatorOnNewOrders(
	e *es.Elevator,
	doorTimer *time.Timer,
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	floor := controller.GetFloor()
	if floor < 0 || floor >= config.NumFloors {
		return
	}
	if e.Behaviour() != es.Idle || e.CurrentFloor() != floor {
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
		requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
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
