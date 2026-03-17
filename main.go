package main

import (
	"elevatorproject/src/callhandler"
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/networking"
	"elevatorproject/src/orderdistributor"
	"flag"
	"fmt"
	"time"
)

func main() {
	port := flag.Int("port", 15657, "Port for the elevator simulator")
	flag.Parse()

	//Set my ID before starting any goroutines.
	config.SetMyID()
	// Initialize channels for communication between goroutines:
	HallOrderUpdate := make(chan elevatorserver.HallOrderUpdate, config.NumFloors*1000) // Buffered to avoid blocking on sends
	CabOrderUpdate := make(chan elevatorserver.CabOrderUpdate, config.NumFloors*1000)
	ElevatorStateUpdate := make(chan elevator.Elevator, config.NumFloors*4)

	PeerUpdate := make(chan []string, 10) // Buffered to avoid blocking on sends

	CurrentOrdersToCallhandler := make(chan elevatorserver.CallHandlerMessage, config.NumFloors*10)
	WorldviewToOrderDistributor := make(chan elevatorserver.OrderDistributorMessage, 5)
	SendWorldviewToNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	ReceiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	ActiveLocalOrders := make(chan [config.NumFloors][config.NumButtons]bool, 5)
	elevatorEvent := make(chan elevator.ElevatorEvent, 5)
	ready := make(chan struct{})

	// Start goroutines:
	fmt.Println("Starting controller")
	go controller.RunController(elevatorEvent, *port)
	time.Sleep(1 * time.Second)
	fmt.Println("Starting callhandler")
	go callhandler.RunCallHandler(ready, elevatorEvent, HallOrderUpdate, CabOrderUpdate, ElevatorStateUpdate, CurrentOrdersToCallhandler, ActiveLocalOrders)
	<-ready
	fmt.Println("Starting elevatorserver")
	go elevatorserver.Run(HallOrderUpdate, CabOrderUpdate, ElevatorStateUpdate, PeerUpdate, CurrentOrdersToCallhandler, WorldviewToOrderDistributor, SendWorldviewToNetwork, ReceiveWorldviewFromNetwork)
	fmt.Println("Starting networking")
	go networking.Run(SendWorldviewToNetwork, PeerUpdate, ReceiveWorldviewFromNetwork)
	fmt.Println("Starting orderdistributor")
	go orderdistributor.Run(WorldviewToOrderDistributor, ActiveLocalOrders)
	fmt.Println("Starting select")
	select {} // Block forever
}
