package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"time"
)

//This file contains the main logic for the elevator server,
// which maintains the latest snapshot of all orders and elevator states.
//It merges incoming updates from both the local elevator and the network using the barrier protocol defined in merge.go,
// ensuring that orders are preserved across failures.
//It periodically publishes the latest merged state to the call handler, order distributor, and networking channels.

// processNetworkMessages receives messages from the network and forwards the
// unpacked orders and elevator state into the local update channels.
func processNetworkMessages(
	channelFromNetworking <-chan NetworkingDistributorMessage,
	hallUpdate chan<- HallOrderUpdate,
	cabUpdate chan<- CabOrderUpdate,
	elevatorStateUpdate chan<- elevator.Elevator,
) {
	for message := range channelFromNetworking {
		tempCab, tempHall, tempElevator := message.UnpackForNetworking()
		newHallOrders := UnpackHallOrders(message.SenderID(), tempHall)
		for _, u := range newHallOrders {
			hallUpdate <- u
		}
		newCabOrders := UnpackCabOrders(tempCab)
		for _, u := range newCabOrders {
			cabUpdate <- u
		}
		elev, ok := tempElevator[message.SenderID()]
		if !ok || elev == nil {
			continue
		}
		elevCopy := elev.Copy()
		elevatorStateUpdate <- elevCopy
	}
}

func copyAllElevState(m map[string]*elevator.Elevator) map[string]*elevator.Elevator {
	cp := make(map[string]*elevator.Elevator, len(m))
	for id, e := range m {
		if e != nil {
			elevCopy := e.Copy()
			cp[id] = &elevCopy
		}
	}
	return cp
}

// publishToConsumers builds messages from the latest snapshots and publishes
// them to the call handler, order distributor, and networking channels.
func publishToConsumers(
	latestHall *orders.HallOrders,
	latestCab map[string]*orders.CabOrders,
	latestElevState map[string]*elevator.Elevator,
	channelForCallHandler chan CallHandlerMessage,
	channelForOrderDistributor chan OrderDistributorMessage,
	channelForNetworking chan NetworkingDistributorMessage,
) {
	var hallValue orders.HallOrders
	if latestHall != nil {
		if cp := latestHall.Copy(); cp != nil {
			hallValue = *cp
		}
	}

	cabValue := make(map[string]orders.CabOrders, len(latestCab))
	for id, cab := range latestCab {
		if cab != nil {
			cabValue[id] = *cab.Copy()
		}
	}

	elevValue := make(map[string]elevator.Elevator, len(latestElevState))
	for id, e := range latestElevState {
		if e != nil {
			elevValue[id] = *e
		}
	}

	myCab, ok := cabValue[config.MyID()]
	if !ok {
		myCab = orders.CabOrders{}
	}

	chMsg := CallHandlerMessage{
		mergedHallOrders: hallValue,
		myCabOrders:      myCab,
	}
	odMsg := OrderDistributorMessage{
		mergedHallOrders: hallValue,
		allCabOrders:     cabValue,
		elevatorState:    elevValue,
	}
	netMsg := NetworkingDistributorMessage{
		senderID:         config.MyID(),
		allCabOrders:     cabValue,
		mergedHallOrders: hallValue,
		elevatorState:    elevValue,
	}
	//Discard outdated snapshots if the channel is full, ensuring the latest state is always published
	// at the next tick without blocking.
	select {
	case <-channelForCallHandler:
	default:
	}
	select {
	case <-channelForOrderDistributor:
	default:
	}
	select {
	case <-channelForNetworking:
	default:
	}

	channelForCallHandler <- chMsg
	channelForOrderDistributor <- odMsg
	channelForNetworking <- netMsg
}

// Run runs the elevator server.
// It listens for local updates to hall orders, cab orders, and elevator state,
// as well as peer updates and incoming messages from the network.
// It maintains the latest snapshot of all orders and elevator states,
// merging incoming updates using the barrier protocol to preserve orders across failures.
// It periodically publishes the latest merged state to the callhandler, orderdistributor, and networking channels.
func Run(
	hallUpdate chan HallOrderUpdate,
	cabUpdate chan CabOrderUpdate,
	elevatorStateUpdate chan elevator.Elevator,
	peersOnlineUpdate <-chan []string,
	channelToCallHandler chan CallHandlerMessage,
	channelToOrderDistributor chan OrderDistributorMessage,
	channelToNetworking chan NetworkingDistributorMessage,
	channelFromNetworking <-chan NetworkingDistributorMessage,
) {
	//For storing the latest snapshot of all orders and elevator states, used for merging and broadcasting to network.
	allHall := map[string]*orders.HallOrders{}
	allCab := map[string]*orders.CabOrders{}
	allElevatorStates := map[string]*elevator.Elevator{}
	elevatorsOnNetwork := []string{}

	// Create your own orders first
	allHall[config.MyID()] = orders.CreateHallOrders()
	allCab[config.MyID()] = orders.CreateCabOrders()
	initialElevatorState := <-elevatorStateUpdate
	allElevatorStates[config.MyID()] = &initialElevatorState

	// Internal snapshot channels: update loop sends latest state, broadcast goroutine reads it.
	hallSnap := make(chan *orders.HallOrders, 1)
	cabSnap := make(chan map[string]*orders.CabOrders, 1)
	elevStateSnap := make(chan map[string]*elevator.Elevator, 1)

	// Throttle: buffer snapshots from the main loop and publish to consumers
	// only at HeartbeatInterval, avoiding flooding
	// on rapid state changes.
	go func() {
		ticker := time.NewTicker(config.HeartbeatInterval)
		defer ticker.Stop()
		latestHall := allHall[config.MyID()]
		latestCab := orders.CopyAllCab(allCab)
		latestElevState := copyAllElevState(allElevatorStates)
		for {
			select {
			case h := <-hallSnap:
				latestHall = h
			case c := <-cabSnap:
				latestCab = c
			case es := <-elevStateSnap:
				latestElevState = es
			case <-ticker.C:
				publishToConsumers(latestHall, latestCab, latestElevState,
					channelToCallHandler, channelToOrderDistributor, channelToNetworking)
			}
		}
	}()
	go processNetworkMessages(channelFromNetworking, hallUpdate, cabUpdate, elevatorStateUpdate)

	for {
		select {
		case u := <-hallUpdate:
			if _, ok := allHall[u.SenderID]; !ok {
				allHall[u.SenderID] = orders.CreateHallOrders()
			}
			allHall[u.SenderID].UpdateOrderState(u.Floor, u.Direction, u.State)
			nextState := mergeHallOrderState(u, config.MyID(), allHall, elevatorsOnNetwork)
			allHall[config.MyID()].UpdateOrderState(u.Floor, u.Direction, nextState)
			// Empty the channel to always have the lates snapshot
			select {
			case <-hallSnap:
			default:
			}
			hallSnap <- allHall[config.MyID()].Copy()

		case u := <-cabUpdate:
			if _, ok := allCab[u.SenderID]; !ok {
				allCab[u.SenderID] = orders.CreateCabOrders()
			}
			allCab[u.SenderID].UpdateOrderState(u.Floor, u.State)
			nextState := mergeCabOrderState(u, allCab, elevatorsOnNetwork)
			allCab[u.SenderID].UpdateOrderState(u.Floor, nextState)
			// Empty the channel to always have the latest snapshot
			select {
			case <-cabSnap:
			default:
			}
			cabSnap <- orders.CopyAllCab(allCab)

		case nodes := <-peersOnlineUpdate:
			elevatorsOnNetwork = nodes
			for _, id := range nodes {
				if _, ok := allHall[id]; !ok {
					allHall[id] = orders.CreateHallOrders()
				}
				if _, ok := allCab[id]; !ok {
					allCab[id] = orders.CreateCabOrders()
				}
			}
		case es := <-elevatorStateUpdate:
			// Always overwrite the elevator state for the sender, since it's a direct report of its physical state.
			allElevatorStates[es.Id()] = &es
			select {
			case <-elevStateSnap:
			default:
			}
			elevStateSnap <- copyAllElevState(allElevatorStates)
		}
	}
}
