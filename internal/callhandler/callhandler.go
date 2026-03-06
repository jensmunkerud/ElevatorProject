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
	es.Initialize("minMac", floorEvent)


	// Create a map for elevators
	elevators := make(map[string]*es.Elevator)
	elev1 := &es.Elevator{}

	// Create and initialize an elevator with dummy data
	elev1.Initialize("bankID", , "down")
	elevators[elev1.getId] = elev1
	ElevatorButtons
	for {
		eb, err := runCostFunc(hallRequests, elevators)
		if err != nil {
			fmt.Printf("Error running costFunc: %e", err)
			continue
		}
	}
}
