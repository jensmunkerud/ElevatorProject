package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"time"
)

// HallOrderUpdate describes an incoming hall order event from another elevator.
type HallOrderUpdate struct {
	SenderID  string
	Floor     int
	Direction int
	State     orders.OrderState
}

// CabOrderUpdate describes an incoming cab order event from another elevator.
type CabOrderUpdate struct {
	SenderID string
	Floor    int
	State    orders.OrderState
}

// HallOrderUpdatesFromNetwork unpacks a received HallOrders snapshot into individual
// HallOrderUpdate values, one per floor per direction, ready to send into hallUpdates.
func HallOrderUpdatesFromNetwork(senderID string, hallOrders *orders.HallOrders) []HallOrderUpdate {
	updates := make([]HallOrderUpdate, 0, config.NumFloors*2)
	for floor := 0; floor < config.NumFloors; floor++ {
		for dir := 0; dir < 2; dir++ {
			updates = append(updates, HallOrderUpdate{
				SenderID:  senderID,
				Floor:     floor,
				Direction: dir,
				State:     hallOrders.GetOrderState(floor, dir),
			})
		}
	}
	return updates
}

// CabOrderUpdatesFromNetwork unpacks a received allCabOrders map into individual
// CabOrderUpdate values, one per elevator per floor, ready to send into cabUpdates.
func CabOrderUpdatesFromNetwork(allCabOrders map[string]*orders.CabOrders) []CabOrderUpdate {
	updates := make([]CabOrderUpdate, 0, len(allCabOrders)*config.NumFloors)
	for elevID, cab := range allCabOrders {
		for floor := 0; floor < config.NumFloors; floor++ {
			updates = append(updates, CabOrderUpdate{
				SenderID: elevID,
				Floor:    floor,
				State:    cab.GetOrderState(floor),
			})
		}
	}
	return updates
}

// mergeHallOrderState resolves the next OrderState for a hall order at the given floor and direction.
// It compares the incoming state from update.SenderID against the receiver's local state, using the
// barrier protocol to coordinate transitions across all online nodes.
func mergeHallOrderState(update HallOrderUpdate, receiverID string, allOrders map[string]*orders.HallOrders, onlineNodes []string) orders.OrderState {
	local := allOrders[receiverID].GetOrderState(update.Floor, update.Direction)
	return mergeState(update.State, local, onlineNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allOrders[id]
		if !ok {
			return orders.UnknownOrderState, false
		}
		return elev.GetOrderState(update.Floor, update.Direction), true
	})
}

// mergeCabOrderState resolves the next OrderState for a cab order at the given floor.
// It uses ownerID (the elevator whose cab button was pressed) as the sole barrier node,
// since cab orders are per-elevator and there is no shared physical button to reach
// cross-node consensus on. The barrier advances once the receiver's local knowledge of
// the owner's state reaches the threshold.
func mergeCabOrderState(update CabOrderUpdate, allCabOrders map[string]*orders.CabOrders, onlineNodes []string) orders.OrderState {
	local := allCabOrders[update.SenderID].GetOrderState(update.Floor)
	return mergeState(update.State, local, onlineNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allCabOrders[id]
		if !ok {
			return orders.UnknownOrderState, false
		}
		return elev.GetOrderState(update.Floor), true
	})
}

// mergeState contains the shared state-machine logic for merging an incoming order state with
// the local state. It applies two distributed barriers: Unconfirmed→Confirmed (all nodes must
// have seen the order) and Completed→Removed (all nodes must have served it). getState is a
// callback that retrieves a node's current OrderState by ID, returning false if the node is unknown.
func mergeState(newOrder orders.OrderState, local orders.OrderState, onlineNodes []string, getState func(string) (orders.OrderState, bool)) orders.OrderState {
	// Unknown always loses
	if local == orders.UnknownOrderState {
		return newOrder
	}
	if newOrder == orders.UnknownOrderState {
		return local
	}

	// Completed is transient — Removed + Completed resets
	if local == orders.RemovedOrderState && newOrder == orders.CompletedOrderState {
		return orders.RemovedOrderState
	}

	// Unconfirmed + Completed resets (missed everything)
	if local == orders.UnconfirmedOrderState && newOrder == orders.CompletedOrderState {
		return orders.RemovedOrderState
	}

	// Barrier 1: Unconfirmed → Confirmed
	if local == orders.UnconfirmedOrderState && newOrder == orders.UnconfirmedOrderState {
		if barrierReached(onlineNodes, orders.UnconfirmedOrderState, getState) {
			return orders.ConfirmedOrderState
		}
		return orders.UnconfirmedOrderState
	}

	// Barrier 2: Completed → Removed
	if local == orders.CompletedOrderState {
		if barrierReached(onlineNodes, orders.CompletedOrderState, getState) {
			return orders.RemovedOrderState
		}
		return orders.CompletedOrderState
	}

	// Default: highest state wins
	if newOrder > local {
		return newOrder
	}
	return local
}

// barrierReached returns true if every online node has an OrderState at or above threshold.
// Returns false immediately if any node is missing from the order map or has not yet reached the threshold.
func barrierReached(onlineNodes []string, threshold orders.OrderState, getState func(string) (orders.OrderState, bool)) bool {
	for _, id := range onlineNodes {
		state, ok := getState(id)
		if !ok {
			return false
		}
		if state < threshold {
			return false
		}
	}
	return true
}

type CallHandlerMessage struct {
	mergedHallOrders orders.HallOrders
	myCabOrders      orders.CabOrders
}

// NewCallHandlerMessage constructs a CallHandlerMessage from HallOrders and CabOrders snapshots.
func NewCallHandlerMessage(hall *orders.HallOrders, cab *orders.CabOrders) CallHandlerMessage {
	var merged orders.HallOrders
	var myCab orders.CabOrders

	if hall != nil {
		if cp := hall.Copy(); cp != nil {
			merged = *cp
		}
	}

	if cab != nil {
		if cp := cab.Copy(); cp != nil {
			myCab = *cp
		}
	}

	return CallHandlerMessage{
		mergedHallOrders: merged,
		myCabOrders:      myCab,
	}
}

// UnpackForCallHandler returns pointer-based snapshots for call handler consumers.
func (m CallHandlerMessage) UnpackForCallHandler() (*orders.HallOrders, *orders.CabOrders) {
	hallOrders := m.mergedHallOrders.Copy()
	cabOrders := m.myCabOrders.Copy()
	return hallOrders, cabOrders
}

type OrderDistributorMessage struct {
	mergedHallOrders orders.HallOrders
	allCabOrders     map[string]orders.CabOrders
	elevatorState    map[string]elevator.Elevator
}

// UnpackForOrderDistributor returns pointer-based snapshots for order distributor consumers.
func (m OrderDistributorMessage) UnpackForOrderDistributor() (map[string]*orders.CabOrders, *orders.HallOrders, map[string]*elevator.Elevator) {
	allCabOrders := make(map[string]*orders.CabOrders, len(m.allCabOrders))
	for id, cab := range m.allCabOrders {
		allCabOrders[id] = cab.Copy()
	}

	hallOrders := m.mergedHallOrders.Copy()

	elevatorState := make(map[string]*elevator.Elevator, len(m.elevatorState))
	for id, elev := range m.elevatorState {
		elevCopy := elev
		elevatorState[id] = &elevCopy
	}

	return allCabOrders, hallOrders, elevatorState
}

type NetworkingDistributorMessage struct {
	// Consider changing order to keep consistency. Note any
	// errors that may arise.
	senderID         string
	allCabOrders     map[string]orders.CabOrders
	mergedHallOrders orders.HallOrders
	elevatorState    map[string]elevator.Elevator
}

func (m *NetworkingDistributorMessage) SenderID() string {
	return m.senderID
}

func NewNetworkingDistributorMessage(
	allCabOrders map[string]*orders.CabOrders,
	hallOrders *orders.HallOrders,
	elevatorState map[string]*elevator.Elevator,
) NetworkingDistributorMessage {
	msg := NetworkingDistributorMessage{
		allCabOrders:  make(map[string]orders.CabOrders, len(allCabOrders)),
		elevatorState: make(map[string]elevator.Elevator, len(elevatorState)),
	}
	if hallOrders != nil {
		msg.mergedHallOrders = *hallOrders.Copy()
	}
	for id, cab := range allCabOrders {
		if cab == nil {
			continue
		}
		msg.allCabOrders[id] = *cab.Copy()
	}
	for id, elev := range elevatorState {
		if elev == nil {
			continue
		}
		msg.elevatorState[id] = *elev
	}
	return msg
}

// UnpackForNetworking returns pointer-based snapshots for networking consumers.
func (m NetworkingDistributorMessage) UnpackForNetworking() (map[string]*orders.CabOrders, *orders.HallOrders, map[string]*elevator.Elevator) {
	allCabOrders := make(map[string]*orders.CabOrders, len(m.allCabOrders))
	for id, cab := range m.allCabOrders {
		allCabOrders[id] = cab.Copy()
	}

	hallOrders := m.mergedHallOrders.Copy()

	elevatorState := make(map[string]*elevator.Elevator, len(m.elevatorState))
	for id, elev := range m.elevatorState {
		elevCopy := elev
		elevatorState[id] = &elevCopy
	}

	return allCabOrders, hallOrders, elevatorState
}

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
		newHallOrders := HallOrderUpdatesFromNetwork(message.SenderID(), tempHall)
		for _, u := range newHallOrders {
			hallUpdate <- u
		}
		newCabOrders := CabOrderUpdatesFromNetwork(tempCab)
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
	//Discard outdated snapshots if the channel is full, ensuring the latest state is always published at the next tick without blocking.
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

// RunElevatorServer runs the elevator server.
// hallOut carries the receiver's updated hall orders at a fixed broadcast interval.
// cabOut carries a snapshot of all known elevators' cab orders each interval, so peers
// can restore orders for an elevator that has gone offline.
func RunElevatorServer(
	hallUpdate chan HallOrderUpdate,
	cabUpdate chan CabOrderUpdate,
	elevatorStateUpdate chan elevator.Elevator,
	peersUpdate <-chan []string,
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

		case nodes := <-peersUpdate:
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
