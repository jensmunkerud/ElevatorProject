package networking

import (
	elevatorstruct "elevatorproject/internal/elevatorStruct"
	"flag"
	"fmt"
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"os"
)

func communicationSetup(currentElevator *elevatorstruct.Elevator) (
	chan peers.PeerUpdate,
	chan bool,
	chan elevatorstruct.Elevator,
	chan elevatorstruct.Elevator) {
	//Create an id for our communication
	udpID := fmt.Itoa(currentElevator.id)
	flag.StringVar(&udpID, "id", "", "id of this peer")
	flag.Parse()

	if udpID == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		udpID = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	peerUpdateChannel := make(chan peers.PeerUpdate)
	enablePeer := make(chan bool)

	go peers.Transmitter(15647, udpID, enablePeer)
	go peers.Receiver(15647, peerUpdateChannel)

	recieveCustomDataType := make(chan elevatorstruct.Elevator)
	sendCustomDataType := make(chan elevatorstruct.Elevator)

	go bcast.Transmitter(16569, sendCustomDataType)
	go bcast.Receiver(16569, recieveCustomDataType)

	return peerUpdateChannel, enablePeer, recieveCustomDataType, sendCustomDataType
}
