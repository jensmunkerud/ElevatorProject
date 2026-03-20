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

	// Decides whether to run in simulator mode or on real hardware.
	simulatorMode := false

	config.InitConfig(*port, simulatorMode)

	if *backup {
		processpair.Run(*port, *masterPID)
		// Master died — we promote to master, fall through to run elevator
	}

	// Spawn and monitor a backup process in the background
	go processpair.SpawnAndMonitorBackup(*port)

	// Initialize channels for communication between goroutines:
	hallOrderUpdate := make(chan elevatorserver.HallOrderUpdate, config.NumFloors*1000) // Buffered to avoid blocking on sends
	cabOrderUpdate := make(chan elevatorserver.CabOrderUpdate, config.NumFloors*1000)
	elevatorStateUpdate := make(chan elevator.Elevator, config.NumFloors*4)

	peerUpdate := make(chan []string, 10) // Buffered to avoid blocking on sends

	ordersOnNetwork := make(chan elevatorserver.CallHandlerMessage, config.NumFloors*10)
	worldviewToOrderDistributor := make(chan elevatorserver.OrderDistributorMessage, 5)
	sendWorldviewToNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	receiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	activeLocalOrders := make(chan [config.NumFloors][config.NumButtons]bool, 5)
	elevatorEvent := make(chan elevator.ElevatorEvent, 5)
	ready := make(chan struct{})

	// Start goroutines:
	fmt.Println("Starting controller")
	go controller.Run(elevatorEvent, *port)
	time.Sleep(1 * time.Second)
	fmt.Println("Starting callhandler")
	go callhandler.Run(ready, elevatorEvent, hallOrderUpdate, cabOrderUpdate, elevatorStateUpdate, ordersOnNetwork, activeLocalOrders)
	<-ready
	fmt.Println("Starting elevatorserver")
	go elevatorserver.Run(hallOrderUpdate, cabOrderUpdate, elevatorStateUpdate, peerUpdate, ordersOnNetwork, worldviewToOrderDistributor, sendWorldviewToNetwork, receiveWorldviewFromNetwork)
	fmt.Println("Starting networking")
	go networking.Run(sendWorldviewToNetwork, peerUpdate, receiveWorldviewFromNetwork)
	fmt.Println("Starting orderdistributor")
	go orderdistributor.Run(worldviewToOrderDistributor, activeLocalOrders)
	fmt.Println("Starting select")
	select {} // Block forever
}
