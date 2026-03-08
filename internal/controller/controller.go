package controller

import (
	"driver-go/elevio"
	"fmt"
	"time"
)

type ElevatorEvent struct {
	OrderEvent       chan elevio.ButtonEvent
	FloorEvent       chan int
	ObstructionEvent chan bool
	StopEvent        chan bool
}

func InitController(ready chan struct{}) *ElevatorEvent {
	orderEvent, floorEvent, obstructionEvent, stopEvent := initElevatorIO(4)

	_, err := initFloor(floorEvent)
	if err != nil {
		fmt.Printf("Error initilizing floor: %d", err)
	}

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

// Checks if we are at a floor and if not, start going downwards to find our first valid floor
func initFloor(floorEvent chan int) (int, error) {
	currentFloor := elevio.GetFloor()
	if currentFloor != -1 {
		return currentFloor, nil
	}
	elevio.SetMotorDirection(elevio.MD_Down)
	currentFloor = <-floorEvent
	elevio.SetMotorDirection(elevio.MD_Stop)
	return currentFloor, nil
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
