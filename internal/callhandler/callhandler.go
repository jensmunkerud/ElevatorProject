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
	c, floor := controller.InitController(ready)
	<-ready

	id, err := getMacAddr()
	if err != nil {
		fmt.Printf("Error finding MAC address")
		return
	}

	localElevator := es.CreateElevator(id, floor, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	updateElevatorState(localElevator)

	for {
		select {
		case order := <-c.OrderEvent:
			fmt.Printf("%+v\n", order)
			// CREATE AND SEND ORDER TO ELEVATORSERVER

		case floor := <-c.FloorEvent:
			fmt.Printf("%+v\n", floor)
			localElevator.Floor = floor
			// CHECK IF ORDER AT FLOOR
			updateElevatorState(localElevator)

		case obstruction := <-c.ObstructionEvent:
			localElevator.b
			fmt.Printf("%+v\n", obstruction)

		case stop := <-c.StopEvent:
			fmt.Printf("%+v\n", stop)

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

func updateElevatorState(localElevator *es.Elevator) {
	switch localElevator.Behaviour() {
	case es.Idle:
		elevio.SetDoorOpenLamp(false)
		break
	case es.MovingUp:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Up)
		break
	case es.MovingDown:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Down)
		break
	case es.DoorOpen:
		elevio.SetDoorOpenLamp(true)
	}
}
