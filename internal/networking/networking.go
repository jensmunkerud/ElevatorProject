package networking

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	elevatorstruct "elevatorproject/internal/elevatorStruct"
	"flag"
)

// Initializes the UDP communication between the elevatorservers
// Returns a channel for peer updates, a channel to enable/disable the transmitter, 
// a channel to receive custom data types and a channel to send custom data types

func communicationSetup(elev *elevatorstruct.Elevator) (
	chan peers.PeerUpdate,
	chan bool,
	chan elevatorstruct.Elevator,
	chan elevatorstruct.Elevator) {

	//Create an id for our communication
	udpID := elev.CurrentElevatorId()
	flag.StringVar(&udpID, "id", "", "id of this peer")
	flag.Parse()

	// Create channel to send and recieve update on names of peers on the network
	peerUpdateChannel := make(chan peers.PeerUpdate)

	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	enablePeer := make(chan bool)

	go peers.Transmitter(15647, udpID, enablePeer)
	go peers.Receiver(15647, peerUpdateChannel)

	recieveCustomDataType := make(chan elevatorstruct.Elevator)
	sendCustomDataType := make(chan elevatorstruct.Elevator)

	go bcast.Transmitter(16568, sendCustomDataType)
	go bcast.Receiver(16569, recieveCustomDataType)

	return peerUpdateChannel, enablePeer, recieveCustomDataType, sendCustomDataType
}
