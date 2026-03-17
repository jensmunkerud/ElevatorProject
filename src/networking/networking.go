package networking

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	es "elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"time"
)

// Run bridges the elevator server with the UDP broadcast network.
// Input: worldview snapshots are serialized into wire Messages and broadcast to peers.
// Output: received worldview snapshots and peer discovery updates are forwarded
// to the provided output channels.
// Ports are defined in config.go.
func Run(
	sendWorldviewIn <-chan es.NetworkingDistributorMessage,
	peerUpdates chan<- []string,
	receiveWorldviewOut chan<- es.NetworkingDistributorMessage,
) {
	peerUpdateCh := make(chan peers.PeerUpdate)
	enablePeer := make(chan bool)
	sendMsg := make(chan Message, 1)
	recvMsg := make(chan Message, 10)

	go peers.Transmitter(config.PeersPort, config.MyID(), enablePeer)

	go func() {
		enablePeer <- true
		time.Sleep(200 * time.Millisecond)
	}()

	go peers.Receiver(config.PeersPort, peerUpdateCh)
	go bcast.Transmitter(config.BroadcastPort, sendMsg)
	go bcast.Receiver(config.BroadcastPort, recvMsg)

	fmt.Println("Starting broadcast loop")
	go func() {
		for worldview := range sendWorldviewIn {
			allCabOrders, hallOrders, elevatorStates := worldview.UnpackForNetworking()
			msg := messageFromWorldview(config.MyID(), hallOrders, allCabOrders, elevatorStates)
			select {
			case sendMsg <- msg:
			default:
				<-sendMsg
				sendMsg <- msg
			}
		}
	}()

	// Inbound loop: decode received messages into order and peer updates.
	fmt.Println("Starting inbound loop")
	for {
		select {
		case msg := <-recvMsg:
			if msg.SenderID == config.MyID() {
				continue
			}
			receiveWorldviewOut <- worldviewFromMessage(msg)

		case pu := <-peerUpdateCh:
			peerUpdates <- pu.Peers
		}
	}
}

func worldviewFromMessage(msg Message) es.NetworkingDistributorMessage {
	hallOrders := orders.CreateHallOrders()
	for floorIdx, floorOrders := range msg.HallOrders {
		for dirIdx, orderState := range floorOrders {
			hallOrders.UpdateOrderState(floorIdx, dirIdx, orders.OrderState(orderState))
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
		elevatorStates[id] = elevator.CreateElevator(
			id,
			state.Floor,
			directionFromString(state.Direction),
			behaviourFromString(state.Behaviour),
		)
	}

	return es.NewNetworkingDistributorMessage(msg.SenderID, allCabOrders, hallOrders, elevatorStates)
}

func directionFromString(direction string) elevator.Direction {
	switch direction {
	case "up":
		return elevator.Up
	case "down":
		return elevator.Down
	default:
		return elevator.Stop
	}
}

func behaviourFromString(behaviour string) elevator.Behaviour {
	switch behaviour {
	case "moving":
		return elevator.Moving
	case "doorOpen":
		return elevator.DoorOpen
	default:
		return elevator.Idle
	}
}
