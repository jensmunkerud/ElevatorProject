package main

import (
	"elevatorproject/src/callhandler"
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/networking"
	"elevatorproject/src/orderdistributor"
	"elevatorproject/src/processpair"
	"flag"
	"fmt"
	"time"
)

func main() {
	port := flag.Int("port", config.ElevatorPort, "Port for the elevator simulator")
	backup := flag.Bool("processpair", false, "Run as backup process that monitors and restarts the elevator")
	masterPID := flag.Int("masterpid", 0, "PID of the master process (used by backup)")
	flag.Parse()

	config.InitConfig(*port)

	if *backup {
		processpair.Run(*port, *masterPID)
		// Master died — we promote to master, fall through to run elevator
	}

	// Spawn and monitor a backup process in the background
	go processpair.SpawnAndMonitorBackup(*port)

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
