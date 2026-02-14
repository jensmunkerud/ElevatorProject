package controller

import (
	"driver-go/elevio"
	"fmt"
	"testing"
)

func TestController(t *testing.T) {
	orderEvent, floorEvent, obstructionEvent, stopEvent := InitController(4)
	myFloor := -1
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
