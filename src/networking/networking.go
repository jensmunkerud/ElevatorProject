package networking

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
)

// Run starts UDP broadcasting for the elevator network. It listens for incoming worldviews, peers on the network and
// sends out the local worldview.
func Run(
	sendWorldviewIn <-chan elevatorserver.NetworkingDistributorMessage,
	peerUpdates chan<- []string,
	receiveWorldviewOut chan<- elevatorserver.NetworkingDistributorMessage,
) {
	myID := config.MyID()
	peerUpdateCh := make(chan peers.PeerUpdate)
	enablePeer := make(chan bool)
	sendMessage := make(chan Message, 1)
	receiveMessage := make(chan Message, 10)

	go peers.Transmitter(config.PeersPort, myID, enablePeer)

	go peers.Receiver(config.PeersPort, peerUpdateCh)
	go bcast.Transmitter(config.BroadcastPort, sendMessage)
	go bcast.Receiver(config.BroadcastPort, receiveMessage)

	fmt.Println("Starting broadcast loop")
	go func() {
		for worldview := range sendWorldviewIn {
			allCabOrders, hallOrders, elevatorStates := worldview.UnpackForNetworking()
			// We always want to send our worldview to peers
			enablePeer <- true

			msg := messageFromWorldview(myID, hallOrders, allCabOrders, elevatorStates)
			select {
			case sendMessage <- msg:
			default:
				<-sendMessage
				sendMessage <- msg
			}
		}
	}()

	// Inbound loop: decode received messages into order and peer updates.
	for {
		select {
		case msg := <-receiveMessage:
			if msg.SenderID == myID {
				continue
			}
			worldview, err := worldviewFromMessage(msg)
			if err != nil {
				fmt.Printf("Skipping malformed message from %s: %v\n", msg.SenderID, err)
				continue
			}
			receiveWorldviewOut <- worldview

		case pu := <-peerUpdateCh:
			peerUpdates <- pu.Peers
		}
	}
}

// Converts a received message from the networking into the internal format used by the elevator server.
// Returns an error if any elevator state fails to parse.
func worldviewFromMessage(msg Message) (elevatorserver.NetworkingDistributorMessage, error) {
	hallOrders := orders.CreateHallOrders()
	for floor, floorOrders := range msg.HallOrders {
		for orderInDirection, orderState := range floorOrders {
			hallOrders.UpdateOrderState(floor, elevator.ConvertOrderType(orderInDirection), orders.OrderState(orderState))
		}
	}

	allCabOrders := make(map[string]*orders.CabOrders, len(msg.AllCabOrders))
	for id, cabOrderStates := range msg.AllCabOrders {
		cab := orders.CreateCabOrders()
		for floor, state := range cabOrderStates {
			cab.UpdateOrderState(floor, orders.OrderState(state))
		}
		allCabOrders[id] = cab
	}

	elevatorStates := make(map[string]*elevator.Elevator, len(msg.ElevatorStates))
	for id, state := range msg.ElevatorStates {
		direction, err := elevator.DirectionFromString(state.Direction)
		if err != nil {
			return elevatorserver.NetworkingDistributorMessage{}, err
		}

		behaviour, err := elevator.BehaviourFromString(state.Behaviour)
		if err != nil {
			return elevatorserver.NetworkingDistributorMessage{}, err
		}

		elev := elevator.CreateElevator(
			id,
			state.Floor,
			direction,
			behaviour,
		)
		elev.UpdateInService(state.InService)
		elevatorStates[id] = elev
	}

	return elevatorserver.NewNetworkingDistributorMessage(msg.SenderID, allCabOrders, hallOrders, elevatorStates), nil
}
