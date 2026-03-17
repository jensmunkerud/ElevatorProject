package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"fmt"
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
		myID := config.MyID()
		newCabOrders := UnpackCabOrders(tempCab)
		for _, u := range newCabOrders {
			if u.SenderID == myID {
				continue // don't let remote views overwrite our own cab orders
			}
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
	fmt.Println("Check 1")
	myID := config.MyID()
	allHall := map[string]*orders.HallOrders{}
	allCab := map[string]*orders.CabOrders{}
	allElevatorStates := map[string]*elevator.Elevator{}
	elevatorsOnNetwork := []string{}
	fmt.Println("Check 2")
	// Create your own orders first
	allHall[myID] = orders.CreateHallOrders()
	allCab[myID] = orders.CreateCabOrders()
	fmt.Println("Check 2.5")
	initialElevatorState := <-elevatorStateUpdate
	fmt.Println("Check 2.6")
	allElevatorStates[myID] = &initialElevatorState
	fmt.Println("Check 3")
	// Internal snapshot channels: update loop sends latest state, broadcast goroutine reads it.
	hallSnap := make(chan *orders.HallOrders, 1)
	cabSnap := make(chan map[string]*orders.CabOrders, 1)
	elevStateSnap := make(chan map[string]*elevator.Elevator, 1)
	fmt.Println("Check 4")
	// Throttle: buffer snapshots from the main loop and publish to consumers
	// only at HeartbeatInterval, avoiding flooding
	// on rapid state changes.
	fmt.Println("Starting publish to consumers loop")
	go func() {
		ticker := time.NewTicker(config.HeartbeatInterval)
		defer ticker.Stop()
		latestHall := allHall[myID]
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
	fmt.Println("Starting process network messages loop")
	go processNetworkMessages(channelFromNetworking, hallUpdate, cabUpdate, elevatorStateUpdate)

	fmt.Println("Starting hallupdate loop")
	for {
		select {
		case u := <-hallUpdate:
			applyHallUpdate(u, myID, allHall, elevatorsOnNetwork)
			select {
			case <-hallSnap:
			default:
			}
			hallSnap <- allHall[myID].Copy()

		case u := <-cabUpdate:
			applyCabUpdate(u, allCab, elevatorsOnNetwork)
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
			id := es.Id()
			allElevatorStates[id] = &es
			select {
			case <-elevStateSnap:
			default:
			}
			elevStateSnap <- copyAllElevState(allElevatorStates)
		}
	}
}

func applyHallUpdate(u HallOrderUpdate, myID string, allHall map[string]*orders.HallOrders, onlineNodes []string) {
	if _, ok := allHall[u.SenderID]; !ok {
		allHall[u.SenderID] = orders.CreateHallOrders()
	}
	allHall[u.SenderID].UpdateOrderState(u.Floor, u.Direction, u.State)
	nextState := mergeHallOrderState(u, myID, allHall, onlineNodes)
	allHall[myID].UpdateOrderState(u.Floor, u.Direction, nextState)
}

func applyCabUpdate(u CabOrderUpdate, allCab map[string]*orders.CabOrders, onlineNodes []string) {
	if _, ok := allCab[u.SenderID]; !ok {
		allCab[u.SenderID] = orders.CreateCabOrders()
	}
	allCab[u.SenderID].UpdateOrderState(u.Floor, u.State)
	nextState := mergeCabOrderState(u, allCab, onlineNodes)
	allCab[u.SenderID].UpdateOrderState(u.Floor, nextState)
}
