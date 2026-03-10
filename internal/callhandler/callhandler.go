package callhandler

// In:	ElevatorEvent chan [controller], cost func chan [costfunc]
// Out:	ElevatorServer.RequestOrder(order data) [elevatorServer], controller.setMotorDirection [controller]

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
			// ElevatorServer.RequestOrder(order data)
			// -> Requests elevatorserver to actually create (or not) a new order,
			// callHandler does not have this authority.

		case floor := <-c.FloorEvent:
			fmt.Printf("%+v\n", floor)
			localElevator.UpdateCurrentFloor(floor)
			// CHECK IF ORDER AT FLOOR
			updateElevatorState(localElevator)

		case obstruction := <-c.ObstructionEvent:
			fmt.Printf("%+v\n", obstruction)
			if obstruction {
				localElevator.UpdateBehaviour(es.Idle)
				// localElevator.UpdateCurrentDirection(es.Stop)
				updateElevatorState(localElevator)
			} else {
				localElevator.UpdateBehaviour(es.Moving) // Possibly dangerous?
				updateElevatorState(localElevator)
			}
			break

		case stop := <-c.StopEvent:
			fmt.Printf("%+v\n", stop)
			localElevator.UpdateBehaviour(es.Idle)
			// localElevator.UpdateCurrentDirection(es.Stop)
			break
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
		elevio.SetMotorDirection(elevio.MD_Stop)
		break
	case es.Moving:
		elevio.SetDoorOpenLamp(false)

		switch localElevator.CurrentDirection() {
		case es.Stop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		case es.Up:
			elevio.SetMotorDirection(elevio.MD_Up)
			break
		case es.Down:
			elevio.SetMotorDirection(elevio.MD_Down)
			break
		}

		break
	case es.DoorOpen:
		elevio.SetDoorOpenLamp(true)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}
