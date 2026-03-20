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
)

/*
This program controls N elevators over M floors. N and M are defined in config.go. The program is
based on a peer-to-peer network of elevators using UDP to communicate. It is split into six main components:
1. Controller: "The physical elevator" Takes IO from the elevator-panel and transforms it into usable formats.
2. Callhandler: "The elevator logic" Handles the logic for one ("my") elevator, and shares it's state with Elevatorserver.
3. Elevatorserver: "The database" Merging of orders and elevator states recieved from the network and the local elevator.
4. Networking: "The communicator" Networking between elevators.
5. Orderdistributor: "The order manager" Assigns cab- and hallcalls to "my" elevator.
6.Processpair: Handles the backup process that monitors and restarts the elevator if it crashes.

Optional flags for the program are:
-port <int> - The port for the elevator IO (standard is 15657)
-processpair - Run as backup process that monitors and restarts the elevator
-masterpid <int> - PID of the master process (used by backup to ensure consistent naming)
-simulator - Run in simulator mode (default is production mode)
	Simulator mode uses the process-id as the node ID and enables running multiple elevators on
	the same machine.
	Production mode uses the MAC address as the node ID and runs the elevator on real hardware.
*/

func main() {
	port := flag.Int("port", config.ElevatorPort, "Port for the elevator simulator")
	backup := flag.Bool("processpair", false, "Run as backup process that monitors and restarts the elevator")
	primaryPID := flag.Int("masterpid", 0, "PID of the master process (used by backup)")
	simulatorMode := flag.Bool("simulator", false, "Run in simulator mode")
	flag.Parse()

	// InitConfig must be called first due to setting global variables.
	config.InitConfig(*port, *simulatorMode)

	// If true, the program will monitor primary process and block until it dies.
	if *backup {
		processpair.RunAsBackup(*port, *primaryPID)
	}

	// If this runs, the program has entered into primary mode.
	go processpair.RunAsPrimary(*port)

	// Buffers are needed to avoid deadlocks during initalization.
	hallOrderUpdate := make(chan elevatorserver.HallOrderUpdate, config.NumFloors*config.NumButtons*10)
	cabOrderUpdate := make(chan elevatorserver.CabOrderUpdate, config.NumFloors*config.NumButtons*10)
	elevatorStateUpdate := make(chan elevator.Elevator, config.NumFloors*4)
	peerUpdate := make(chan []string, 10)
	ordersOnNetwork := make(chan elevatorserver.CallHandlerMessage, config.NumFloors*10)
	worldviewToOrderDistributor := make(chan elevatorserver.OrderDistributorMessage, 5)
	sendWorldviewToNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	receiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage, 5)
	activeLocalOrders := make(chan [config.NumFloors][config.NumButtons]bool, 5)
	hardwareEvent := make(chan elevator.HardwareEvent, 5)
	readyCallhandler := make(chan struct{})

	// Start goroutines:
	fmt.Println("Starting controller")
	controller.Run(hardwareEvent, *port)
	fmt.Println("Starting callhandler")
	go callhandler.Run(readyCallhandler, hardwareEvent, hallOrderUpdate, cabOrderUpdate, elevatorStateUpdate, ordersOnNetwork, activeLocalOrders)
	<-readyCallhandler
	fmt.Println("Starting elevatorserver")
	go elevatorserver.Run(hallOrderUpdate, cabOrderUpdate, elevatorStateUpdate, peerUpdate, ordersOnNetwork, worldviewToOrderDistributor, sendWorldviewToNetwork, receiveWorldviewFromNetwork)
	fmt.Println("Starting networking")
	go networking.Run(sendWorldviewToNetwork, peerUpdate, receiveWorldviewFromNetwork)
	fmt.Println("Starting orderdistributor")
	go orderdistributor.Run(worldviewToOrderDistributor, activeLocalOrders)
	// Block forever to keep the program running
	select {}
}
