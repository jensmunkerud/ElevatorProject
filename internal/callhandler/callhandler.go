package callhandler

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	"driver-go/elevio"
	controller "elevatorproject/internal/controller"
	es "elevatorproject/internal/elevatorstruct"
	"fmt"
	"net"
	"testing"
)

func TestCallHandler(t *testing.T) {
	go InitCallHandler()
	select {}
}

func InitCallHandler() {
	ready := make(chan struct{})
	c := InitController(ready)
	<-ready

	elevators := make(map[string]*es.Elevator)

	id, err := getMacAddr()
	if err != nil {
		fmt.Printf("Error finding MAC address")
		return
	}

	localElevator := CreateElevator(id, , es.Direction.Stop, es.Behaviour.Idle)
	elevators[localElevator.Id()] = localElevator

	for {
		select {
		case a := <-orderEvent:
			fmt.Printf("%+v\n", a)
			targetFloor = a.Floor
			if MyFloor < 0 || MyFloor == a.Floor {
				continue
			} else if MyFloor < targetFloor {
				elevio.SetMotorDirection(elevio.MD_Up)
				IsAtFloor = false
			} else if MyFloor > targetFloor {
				elevio.SetMotorDirection(elevio.MD_Down)
				IsAtFloor = false
			}

		case floor := <-floorEvent:
			MyFloor = floor
			if MyFloor == targetFloor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				close(targetDone)
			}

		case a := <-obstructionEvent:
			fmt.Printf("%+v\n", a)

		case a := <-stopEvent:
			fmt.Printf("%+v\n", a)
		}
	}

	for {
		select {
		case a := <-controller.Controller.OrderEvent:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}

	for {
		eb, err := runCostFunc(elevators)
		if err != nil {
			fmt.Printf("Error running costFunc: %e", err)
			continue
		}
	}
}

func getMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a, nil
		}
	}

	return "", fmt.Errorf("no MAC address found")
}
