package controller

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"fmt"
)

func Run(hardwareEvent chan elevator.HardwareEvent, port int) {
	// Initializes external library for communicating with the elevator.
	elevio.Init(fmt.Sprintf("localhost:%d", port), config.NumFloors)

	incomingOrderEvent := make(chan elevio.ButtonEvent)
	outgoingOrderEvent := make(chan elevator.OrderEvent)
	floorEvent := make(chan int)
	obstructionEvent := make(chan bool)
	stopEvent := make(chan bool)

	// Merge all events into a single channel to simplify the main event loop in callhandler.
	hardwareEvent <- elevator.CreateHardwareEvent(outgoingOrderEvent, floorEvent, obstructionEvent, stopEvent)

	go elevio.PollButtons(incomingOrderEvent)
	go elevio.PollFloorSensor(floorEvent)
	go elevio.PollObstructionSwitch(obstructionEvent)
	go elevio.PollStopButton(stopEvent)

	// Converts elevio.ButtonEvents to elevator.OrderEvent to decouple the controller from the elevio package.
	go func() {
		for order := range incomingOrderEvent {
			outgoingOrderEvent <- elevator.CreateOrderEvent(order.Floor, elevator.OrderType(order.Button))
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

func SetButtonLamp(button elevator.OrderType, floor int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, value)
}

func SetDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func SetStopLamp(value bool) {
	elevio.SetStopLamp(value)
}

func GetCurrentFloor() int {
	return elevio.GetFloor()
}

func SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}
