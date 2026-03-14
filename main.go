package main

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
)

func main() {
	//Set my ID before starting any goroutines.
	config.SetMyID()
	// Initialize channels for communication between goroutines:
	hallOrderUpdate := make(chan elevatorserver.HallOrderUpdate, config.NumFloors*10) // Buffered to avoid blocking on sends
	cabOrderUpdate := make(chan elevatorserver.CabOrderUpdate, config.NumFloors*10)
	elevatorStateLocal := make(chan elevator.Elevator, config.NumFloors*4)
	currentOrders := make(chan elevatorserver.CallHandlerMessage, config.NumFloors*10)
	PeerUpdate := make(chan []string, 10) // Buffered to avoid blocking on sends

	WorldviewToOrderDistributor:= make(chan elevatorserver.OrderDistributorMessage, 5)
	SendWorldviewToNetwork:= make(chan elevatorserver.NetworkingDistributorMessage, 5)
	ReceiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)

	ActiveOrders := make(chan [][]bool)

	// Start goroutines:
	// go controller(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal)
	// go callHandler(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, CurrentOrders)
	// go ElevatorServer(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, WorldviewToCostFunction, SendWorldviewToNetwork, ReceiveWorldviewFromNetwork, CurrentOrders)
	// go RunNetworking(SendWorldviewToNetwork, PeerUpdate, ReceiveWorldviewFromNetwork)
	// go costFunction(WorldviewToCostFunction, ActiveOrders)
}