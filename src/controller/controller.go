package controller

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"fmt"
)

func RunController(elevatorEvent chan es.ElevatorEvent, port int) {
	// Initializes communication with elevatorserver to receive IO from physical elevator
	elevio.Init(fmt.Sprintf("localhost:%d", port), config.NumFloors)

	orderEventElevio := make(chan elevio.ButtonEvent)
	orderEvent := make(chan es.ButtonEvent)
	floorEvent := make(chan int)
	obstructionEvent := make(chan bool)
	stopEvent := make(chan bool)

	elevatorEvent <- es.ElevatorEvent{
		OrderEvent:       orderEvent,
		FloorEvent:       floorEvent,
		ObstructionEvent: obstructionEvent,
		StopEvent:        stopEvent,
	}

	// Continually polls hardwarechanges onto the event channels
	go elevio.PollButtons(orderEventElevio)
	go elevio.PollFloorSensor(floorEvent)
	go elevio.PollObstructionSwitch(obstructionEvent)
	go elevio.PollStopButton(stopEvent)

	// Necessary to make elevio and callhandler "loosely coupled":
	// orderEvent simply reflects orderEventElevio, but with internal ButtonEvent type
	fmt.Println("Starting controller loop")
	go func() {
		for order := range orderEventElevio {
			orderEvent <- es.ButtonEvent{
				Floor:  order.Floor,
				Button: es.OrderType(order.Button),
			}
		}
	}()
}

func MoveElevatorUp() {
	if !(elevio.GetFloor() >= config.NumFloors-1) {
		elevio.SetMotorDirection(elevio.MD_Up)
	}
}

func MoveElevatorDown() {
	if !(elevio.GetFloor() == 0) {
		elevio.SetMotorDirection(elevio.MD_Down)
	}
}

func StopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func SetButtonLamp(button es.OrderType, floor int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, value)
}

func SetDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func SetStopLamp(value bool) {
	elevio.SetStopLamp(value)
}

func GetFloor() int {
	return elevio.GetFloor()
}

func SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}
