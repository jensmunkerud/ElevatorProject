package networking

import(
	"testing"
	"elevatorproject/internal/elevatorStruct"
	"time"
	"fmt"
)

func Test(t *testing.T) {
	var elevator elevatorstruct.Elevator
	elevator.Initialize("1", 0, "up")
	peerUpdateChannel, enablePeer, receiveCustomDataType, sendCustomDataType := communicationSetup(&elevator)

	go func() {
		for {
			sendCustomDataType <- elevator
			time.Sleep(1 * time.Second)
		}
	}()
	fmt.Println("Started sending elevator data")
	fmt.Printf("Enable peer %v\n", enablePeer)
	for{
		select{
			case peerUpdate := <-peerUpdateChannel:
				fmt.Printf("Peer update: \n")
				fmt.Printf("Peers : %q\n", peerUpdate.Peers)
				fmt.Printf("New : %q\n", peerUpdate.New)
				fmt.Printf("Lost : %q\n", peerUpdate.Lost)
			case recievedData := <- receiveCustomDataType:
				fmt.Printf("Received elevator data: %#v\n", recievedData)
		}
	}
}