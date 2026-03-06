package controller

import (
	"driver-go/elevio"
	"time"
)

type Controller struct {
	OrderEvent       chan elevio.ButtonEvent
	FloorEvent       chan int
	ObstructionEvent chan bool
	StopEvent        chan bool
	MyFloor          int
	IsAtFloor        bool
	Obstructed       bool
}

func InitController(ready chan struct{}) *Controller {
	orderEvent, floorEvent, obstructionEvent, stopEvent := initElevatorIO(4)

	myFloor := initFloor(floorEvent)
	isAtFloor := true

	c := &Controller{
		OrderEvent:       orderEvent,
		FloorEvent:       floorEvent,
		ObstructionEvent: obstructionEvent,
		StopEvent:        stopEvent,
		MyFloor:          myFloor,
		IsAtFloor:        isAtFloor,
	}

	close(ready)
	return c
	// for {
	// 	select {
	// 	case a := <-orderEvent:
	// 		fmt.Printf("%+v\n", a)
	// 		targetFloor = a.Floor
	// 		if MyFloor < 0 || MyFloor == a.Floor {
	// 			continue
	// 		} else if MyFloor < targetFloor {
	// 			elevio.SetMotorDirection(elevio.MD_Up)
	// 			IsAtFloor = false
	// 		} else if MyFloor > targetFloor {
	// 			elevio.SetMotorDirection(elevio.MD_Down)
	// 			IsAtFloor = false
	// 		}

	// 	case floor := <-floorEvent:
	// 		MyFloor = floor
	// 		if MyFloor == targetFloor {
	// 			elevio.SetMotorDirection(elevio.MD_Stop)
	// 			close(targetDone)
	// 		}

	// 	case a := <-obstructionEvent:
	// 		fmt.Printf("%+v\n", a)

	// 	case a := <-stopEvent:
	// 		fmt.Printf("%+v\n", a)
	// 	}
	// }
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

// Starts off the elevator, first check if we are at a floor and if not startgoing downwards to find our first valid floor
func initFloor(floorEvent chan int) int {
	currentFloor := elevio.GetFloor()
	if currentFloor != -1 {
		return currentFloor
	}
	elevio.SetMotorDirection(elevio.MD_Down)
	currentFloor = <-floorEvent
	elevio.SetMotorDirection(elevio.MD_Stop)
	return currentFloor
}

func (c *Controller) monitorObstruction() {
	for o := range c.ObstructionEvent {
		c.Obstructed = o
	}
}

func (c *Controller) waitUntilClear() {
	for c.Obstructed {
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *Controller) openDoor() {
	timer := time.NewTimer(3 * time.Second)
	elevio.SetDoorOpenLamp(true)
	for {
		select {

		case o := <-c.ObstructionEvent:
			if o {
				timer.Stop()
			} else {
				timer.Reset(3 * time.Second)
			}

		case <-timer.C:
			elevio.SetDoorOpenLamp(false)
			return
		}
	}
}

func (c *Controller) GoToFloor(target int, done chan struct{}) {
	if c.MyFloor == target {
		return
	} else if c.MyFloor < target {
		c.openDoor()
		elevio.SetMotorDirection(elevio.MD_Up)
	} else {
		c.openDoor()
		elevio.SetMotorDirection(elevio.MD_Up)
	}
	for {
		select {
		case floor := <-c.FloorEvent:
			if floor == target {
				elevio.SetMotorDirection(elevio.MD_Stop)
				close(done)
				return
			}
		}
	}
}
