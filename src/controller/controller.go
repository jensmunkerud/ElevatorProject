package controller

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
)

type ElevatorEvent struct {
	OrderEvent       chan es.ButtonType
	FloorEvent       chan int
	ObstructionEvent chan bool
	StopEvent        chan bool
}

func InitController(ready chan struct{}) *ElevatorEvent {
	orderEvent, floorEvent, obstructionEvent, stopEvent := initElevatorIO()
	c := &ElevatorEvent{
		OrderEvent:       orderEvent,
		FloorEvent:       floorEvent,
		ObstructionEvent: obstructionEvent,
		StopEvent:        stopEvent,
	}
	close(ready)
	return c
}

// Initializes communication with elevatorserver to receive IO from physical elevator
func initElevatorIO() (
	chan elevio.ButtonEvent,
	chan int,
	chan bool,
	chan bool,
) {
	elevio.Init("localhost:15657", config.NumFloors)

	orderEventElevio := make(chan elevio.ButtonEvent)
	floorEvent := make(chan int)
	obstructionEvent := make(chan bool)
	stopEvent := make(chan bool)

	go elevio.PollButtons(orderEventElevio)
	go elevio.PollFloorSensor(floorEvent)
	go elevio.PollObstructionSwitch(obstructionEvent)
	go elevio.PollStopButton(stopEvent)

	return orderEventElevio, floorEvent, obstructionEvent, stopEvent
}

func MoveElevatorUp() {
	elevio.SetMotorDirection(elevio.MD_Up)
}

func MoveElevatorDown() {
	elevio.SetMotorDirection(elevio.MD_Down)
}

func StopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func SetButtonLamp(button es.ButtonType, floor int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, value)
}
