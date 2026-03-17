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
				Button: es.ButtonType(order.Button),
			}
		}
	}()
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
