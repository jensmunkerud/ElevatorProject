package controller

import (
	"driver-go/elevio"
	"fmt"
)

func InitController() {
	orderEvent, floorEvent, obstructionEvent, stopEvent := initElevatorIO(4)
	myFloor := initFloor(floorEvent)
	targetFloor := -1
	for {
		select {
		case a := <-orderEvent:
			fmt.Printf("%+v\n", a)
			targetFloor = a.Floor
			if myFloor < 0 || myFloor == a.Floor {
				continue
			} else if myFloor < targetFloor {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if myFloor > targetFloor {
				elevio.SetMotorDirection(elevio.MD_Down)
			}

		case a := <-floorEvent:
			fmt.Printf("%+v\n", a)
			myFloor = a
			if myFloor == targetFloor || targetFloor < 0 {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if myFloor < targetFloor {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if myFloor > targetFloor {
				elevio.SetMotorDirection(elevio.MD_Down)
			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}

		case a := <-obstructionEvent:
			fmt.Printf("%+v\n", a)

		case a := <-stopEvent:
			fmt.Printf("%+v\n", a)
		}
	}
}

// Handles IO communication between software and hardware
func initElevatorIO(numFloors int) (
	chan elevio.ButtonEvent,
	chan int,
	chan bool,
	chan bool,
) {
	elevio.Init("localhost:15657", numFloors)

	orderEvent := make(chan elevio.ButtonEvent)
	floorEvent := make(chan int)
	obstructionEvent := make(chan bool)
	stopEvent := make(chan bool)

	go elevio.PollButtons(orderEvent)
	go elevio.PollFloorSensor(floorEvent)
	go elevio.PollObstructionSwitch(obstructionEvent)
	go elevio.PollStopButton(stopEvent)

	return orderEvent, floorEvent, obstructionEvent, stopEvent
}

// Starts off the elevator, going downwards to find our first valid floor
func initFloor(floorEvent chan int) int {
	// 1. Check if we're already at a floor
	currentFloor := elevio.GetFloor()
	if currentFloor != -1 {
		return currentFloor
	}

	// 2. We are between floors → move down
	elevio.SetMotorDirection(elevio.MD_Down)

	// 3. Wait until PollFloorSensor detects a floor
	floor := <-floorEvent

	// 4. Stop motor
	elevio.SetMotorDirection(elevio.MD_Stop)

	return floor
}
