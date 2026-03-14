package controller

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
)


func InitController(ready chan struct{}) *es.ElevatorEvent {
	orderEvent, floorEvent, obstructionEvent, stopEvent := initElevatorIO()
	c := &es.ElevatorEvent{
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
	chan es.ButtonEvent,
	chan int,
	chan bool,
	chan bool,
) {
	elevio.Init("localhost:15657", config.NumFloors)

	orderEventElevio := make(chan elevio.ButtonEvent)
	orderEvent := make(chan es.ButtonEvent)
	floorEvent := make(chan int)
	obstructionEvent := make(chan bool)
	stopEvent := make(chan bool)

	go elevio.PollButtons(orderEventElevio)
	go elevio.PollFloorSensor(floorEvent)
	go elevio.PollObstructionSwitch(obstructionEvent)
	go elevio.PollStopButton(stopEvent)

	go func() {
		for order := range orderEventElevio {
			orderEvent <- es.ButtonEvent{
				Floor:  order.Floor,
				Button: es.ButtonType(order.Button),
			}
		}
	}()

	return orderEvent, floorEvent, obstructionEvent, stopEvent
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

