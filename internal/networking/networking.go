package networking

import (
	"network-go/network/bcast"
	"network-go/network/localip"
	"network-go/network/peers"
	"flag"
	"os"
	"time"
)


func communicationSetup(elevatorId string, ) (
	chan peers.PeerUpdate, 
	chan bool,
	chan sendCustomDataType,
	chan receiveCustomDataType) {

}

