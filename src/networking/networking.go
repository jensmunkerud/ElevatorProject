package networking

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevatorserver"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
)

// RunNetworking bridges the elevator server with the UDP broadcast network.
// Outbound: hall, cab, and elevator state snapshots are combined into wire Messages
// and broadcast to all peers on port 16569.
// Inbound: received Messages are decoded into HallOrderUpdates, CabOrderUpdates,
// and elevator state updates forwarded into the provided output channels.
// Peer discovery runs on port 15647; the current peer list is forwarded to peerUpdates.
func RunNetworking(
	receiverID string,
	hallOut <-chan *orders.HallOrders,
	cabOut <-chan map[string]*orders.CabOrders,
	elevatorStateIn <-chan elevator.Elevator,
	hallUpdates chan<- es.HallOrderUpdate,
	cabUpdates chan<- es.CabOrderUpdate,
	peerUpdates chan<- []string,
	elevatorStatesOut chan<- ElevatorStateUpdate,
) {
	peerUpdateCh := make(chan peers.PeerUpdate)
	enablePeer := make(chan bool)
	sendMsg := make(chan Message, 1)
	recvMsg := make(chan Message, 10)

	go peers.Transmitter(15647, receiverID, enablePeer)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16569, sendMsg)
	go bcast.Receiver(16569, recvMsg)

	// Outbound goroutine: rebuild and queue a message whenever a new snapshot arrives.
	// A non-blocking send keeps only the latest message in the buffer.
	go func() {
		latestHall := orders.CreateHallOrders()
		latestCab := map[string]*orders.CabOrders{}
		latestElevState := elevator.Elevator{}
		for {
			select {
			case h := <-hallOut:
				latestHall = h
			case c := <-cabOut:
				latestCab = c
			case s := <-elevatorStateIn:
				latestElevState = s
			}
			msg := messageFromOrders(receiverID, latestHall, latestCab, latestElevState)
			select {
			case sendMsg <- msg:
			default:
				<-sendMsg
				sendMsg <- msg
			}
		}
	}()

	// Inbound loop: decode received messages into order and peer updates.
	for {
		select {
		case msg := <-recvMsg:
			if msg.SenderID == receiverID {
				continue
			} 
			hallOrds := orders.CreateHallOrders()
			for floor := 0; floor < config.NumFloors; floor++ {
				for dir := 0; dir < 2; dir++ {
					hallOrds.UpdateOrderState(floor, dir, orders.OrderState(msg.HallOrders[floor][dir]))
				}
			}
			for _, upd := range es.HallOrderUpdatesFromNetwork(msg.SenderID, hallOrds) {
				hallUpdates <- upd
			}

			allCab := make(map[string]*orders.CabOrders, len(msg.AllCabOrders))
			for id, arr := range msg.AllCabOrders {
				cab := orders.CreateCabOrders()
				for floor := 0; floor < config.NumFloors; floor++ {
					cab.UpdateOrderState(floor, orders.OrderState(arr[floor]))
				}
				allCab[id] = cab
			}
			for _, upd := range es.CabOrderUpdatesFromNetwork(allCab) {
				cabUpdates <- upd
			}

			if state, ok := msg.ElevatorStates[msg.SenderID]; ok {
				elevatorStatesOut <- ElevatorStateUpdate{SenderID: msg.SenderID, State: state}
			}

		case pu := <-peerUpdateCh:
			peerUpdates <- pu.Peers
		}
	}
}
