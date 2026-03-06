package callhandler

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
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
	go controller.InitController(ready)
	<-ready
	elevators := make(map[string]*es.Elevator)

	id, err := getMacAddr()
	if err != nil {
		fmt.Printf("Error finding MAC address")
		return
	}

	localElevator := CreateElevator(id, controller.MyFloor, es.Direction.Stop, es.Behaviour.Idle)
	elevators[localElevator.Id()] = localElevator

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
