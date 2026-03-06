package callhandler

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	controller "elevatorproject/internal/controller"
	es "elevatorproject/internal/elevatorStruct"
	"fmt"
)

func Test() {
	fmt.Println("Hello from callHandler")
}

func InitCallHandler() {
	controller.InitController()
	if controller.isAtFloor {

	}
	elevators := make(map[string]*es.Elevator)
	localElevator := &es.Elevator{}
	localElevator.Initialize("minMac", floorEvent)



	// Create and initialize an elevator with dummy data
	localElevator.Initialize("bankID", , "down")
	elevators[elev1.Id()] = elev1
	ElevatorButtons
	for {
		eb, err := runCostFunc(hallRequests, elevators)
		if err != nil {
			fmt.Printf("Error running costFunc: %e", err)
			continue
		}
	}
}
