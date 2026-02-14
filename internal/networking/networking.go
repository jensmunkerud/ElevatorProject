package networking

import (
	"network-go/network/bcast"
	"network-go/network/localip"
	"network-go/network/peers"
	"elevatorproject/internal/elevatorStruct"
	"flag"
	"os"
	"time"
	"fmt"
)


func communicationSetup(currentElevator *elevatorStruct.Elevator) (
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


