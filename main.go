package main

import (
	"elevatorproject/src/callhandler"
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/networking"
	"elevatorproject/src/orderdistributor"
)

func main() {
	//Set my ID before starting any goroutines.
	config.SetMyID()
	// Initialize channels for communication between goroutines:
	HallOrderUpdate := make(chan elevatorserver.HallOrderUpdate, config.NumFloors*10) // Buffered to avoid blocking on sends
	CabOrderUpdate := make(chan elevatorserver.CabOrderUpdate, config.NumFloors*10)
	ElevatorStateUpdate := make(chan elevator.Elevator, config.NumFloors*4)

	PeerUpdate := make(chan []string, 10) // Buffered to avoid blocking on sends

	CurrentOrdersToCallhandler := make(chan elevatorserver.CallHandlerMessage, config.NumFloors*10)
	WorldviewToOrderDistributor := make(chan elevatorserver.OrderDistributorMessage, 5)
	SendWorldviewToNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	ReceiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	ActiveLocalOrders := make(chan [config.NumFloors][config.NumButtons]bool)
	ElevatorStateLocal := make(chan elevator.Elevator)
	elevatorEvent := make(chan elevator.ElevatorEvent)
	CallHandlerMessage := make(chan elevatorserver.CallHandlerMessage)
	ready := make(chan struct{})

	// Start goroutines:
	go controller.RunController(elevatorEvent)
	go callhandler.RunCallHandler(ready, elevatorEvent, HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, CallHandlerMessage, ActiveLocalOrders)
	<-ready
	go elevatorserver.RunElevatorServer(HallOrderUpdate, CabOrderUpdate, ElevatorStateUpdate, PeerUpdate, CurrentOrdersToCallhandler, WorldviewToOrderDistributor, SendWorldviewToNetwork, ReceiveWorldviewFromNetwork)
	go networking.RunNetworking(SendWorldviewToNetwork, PeerUpdate, ReceiveWorldviewFromNetwork)
	go orderdistributor.RunCostFunc(WorldviewToOrderDistributor, ActiveLocalOrders)
}
