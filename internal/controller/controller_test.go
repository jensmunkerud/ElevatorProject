package controller

import (
	"driver-go/elevio"
	"fmt"
	"testing"
)

func TestController(t *testing.T) {
	orderEvent, floorEvent, obstructionEvent, stopEvent := InitController(4)
	myFloor := -1
	for {
		select {
		case a := <-orderEvent:
			fmt.Printf("%+v\n", a)
			if myFloor < 0 || myFloor == a.Floor {
				continue
			} else if myFloor < a.Floor {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if myFloor > a.Floor {
				elevio.SetMotorDirection(elevio.MD_Down)
			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}

		case a := <-floorEvent:
			fmt.Printf("%+v\n", a)
			if myFloor == a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			myFloor = a

		case a := <-obstructionEvent:
			fmt.Printf("%+v\n", a)

		case a := <-stopEvent:
			fmt.Printf("%+v\n", a)
		}
	}
}
