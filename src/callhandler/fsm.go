package callhandler

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"fmt"
	"time"
)

func initializeElevatorToValidFloor(e *elevator.Elevator) {
	if config.NumFloors < 1 {
		fmt.Println("Error: NumFloors must be at least 1")
		return
	}
	floor := controller.GetCurrentFloor()
	if floor != -1 && floor >= 0 && floor < config.NumFloors {
		if e.Behaviour() != elevator.Moving {
			e.UpdateInService(true)
			return
		}
	}
	mid := config.NumFloors / 2

	if e.CurrentFloor() < mid {
		controller.MoveElevatorUp()
		e.UpdateCurrentDirection(elevator.Up)
	} else {
		controller.MoveElevatorDown()
		e.UpdateCurrentDirection(elevator.Down)
	}
	e.UpdateBehaviour(elevator.Moving)
}

func elevatorArrivedAtFloor(
	e *elevator.Elevator,
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
	case elevator.Moving:
		// Necessary to restart the serviceTimer here, to also mark as out of service if door is obstructed.
		restartTimer(serviceTimer, config.ServiceTimeout)
		if shouldStop(*e) {
			controller.StopElevator()
			controller.SetDoorOpenLamp(true)
			requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
			restartTimer(doorTimer, config.DoorOpenDuration)
			e.UpdateBehaviour(elevator.DoorOpen)
		}
	default:
		// nothing
	}
}

func elevatorOnDoorTimeout(
	e *elevator.Elevator,
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
	case elevator.DoorOpen:
		newDirection, newBehaviour := chooseNextDirection(*e)
		e.UpdateCurrentDirection(newDirection)
		e.UpdateBehaviour(newBehaviour)

		switch e.Behaviour() {
		case elevator.DoorOpen:
			e.UpdateInService(true)
			restartTimer(doorTimer, config.DoorOpenDuration)
			restartTimer(serviceTimer, config.ServiceTimeout)
			requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
		case elevator.Moving, elevator.Idle:
			restartTimer(serviceTimer, config.ServiceTimeout)
			e.UpdateInService(true)
			controller.SetDoorOpenLamp(false)
			if !e.StopPressed() && e.Behaviour() == elevator.Moving {
				switch e.CurrentDirection() {
				case elevator.Up:
					controller.MoveElevatorUp()
				case elevator.Down:
					controller.MoveElevatorDown()
				}
			}
		}

	default:
		// nothing
	}
}

func elevatorOnNewOrders(
	e *elevator.Elevator,
	doorTimer *time.Timer,
	serviceTimer *time.Timer,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
) {
	floor := controller.GetCurrentFloor()
	if floor < 0 || floor >= config.NumFloors {
		return
	}
	if e.Behaviour() != elevator.Idle || e.CurrentFloor() != floor {
		return
	}
	newDirection, newBehaviour := chooseNextDirection(*e)
	e.UpdateCurrentDirection(newDirection)
	e.UpdateBehaviour(newBehaviour)
	e.UpdateInService(true)
	switch newBehaviour {
	case elevator.DoorOpen:
		controller.SetDoorOpenLamp(true)
		restartTimer(doorTimer, config.DoorOpenDuration)
		requestClearAtCurrentFloor(e, hallOrderUpdate, cabOrderUpdate)
	case elevator.Moving:
		restartTimer(serviceTimer, config.ServiceTimeout)
		if !e.StopPressed() {
			switch e.CurrentDirection() {
			case elevator.Up:
				controller.MoveElevatorUp()
			case elevator.Down:
				controller.MoveElevatorDown()
			}
		}
	}
}
