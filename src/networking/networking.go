package networking

import (
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	es "elevatorproject/src/elevator"
	"flag"
	"fmt"
	"os"
)

// CommunicationSetup initializes the networking infrastructure for an elevator to participate
// in a distributed system. It establishes peer discovery (heartbeat/discovery on port 15647)
// and broadcast messaging capabilities (port 16569) to enable inter-elevator communication.
// The function returns channels for managing peer connections and sending/receiving messages.
func CommunicationSetup(message Message, currentElevator *es.Elevator) (
	chan peers.PeerUpdate,
	chan bool,
	chan Message,
	chan Message) {

	// Generate a unique identifier for this elevator instance. Attempt to use the configured
	// elevator ID if available, otherwise fall back to a composite ID using local IP and process ID.
	// This ensures each peer is uniquely identifiable across the network.
	udpID := currentElevator.Id()
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

	// Set up peer discovery: channels for notifying when peers join/leave the network
	peerUpdateChannel := make(chan peers.PeerUpdate)
	enablePeer := make(chan bool)

	// Launch peer discovery service on port 15647. This enables the elevator to advertise
	// its presence and detect other elevators on the network.
	go peers.Transmitter(15647, udpID, enablePeer)
	go peers.Receiver(15647, peerUpdateChannel)

	// Set up broadcast messaging channels for custom message passing between elevators
	recieveCustomDataType := make(chan ms.Message)
	sendCustomDataType := make(chan ms.Message)

	// Launch broadcast service on port 16569. This enables elevators to share state
	// and coordinate actions across the distributed system.
	go bcast.Transmitter(16569, sendCustomDataType)
	go bcast.Receiver(16569, recieveCustomDataType)

	return peerUpdateChannel, enablePeer, recieveCustomDataType, sendCustomDataType
}
