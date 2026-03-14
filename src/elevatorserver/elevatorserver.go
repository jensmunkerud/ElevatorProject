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
// SenderID is set to the owning elevator's ID, not the relaying node's ID.
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

// UnpackForConvertToJson returns pointer-based snapshots compatible with orderdistributor.ConvertToJson.
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

// Takes in the results of the merging of orders and distributes it to
// any packages that may need it.
func distributeResultsToUsers(
	hallOut <-chan *orders.HallOrders,
	cabOut <-chan map[string]*orders.CabOrders,
	elevatorState <-chan map[string]*elevator.Elevator,
	channelForCallHandler chan CallHandlerMessage,
	channelForOrderDistributor chan OrderDistributorMessage,
	channelForNetworking chan NetworkingDistributorMessage,
) () {

	// Latest-only outputs (buffer size 1): never block the distributor goroutine.
	callHandlerOutput := make(chan CallHandlerMessage, 1)
	orderDistributorOutput := make(chan OrderDistributorMessage, 1)
	networkingDistributorOutput := make(chan NetworkingDistributorMessage, 1)

	// Helpers to deep-copy pointer snapshots into value-based message fields.
	copyHallValue := func(h *orders.HallOrders) (orders.HallOrders, bool) {
		if h == nil {
			return orders.HallOrders{}, false
		}
		cp := h.Copy()
		return *cp, true
	}
	copyAllCabValue := func(m map[string]*orders.CabOrders) (map[string]orders.CabOrders, bool) {
		if m == nil {
			return nil, false
		}
		cp := make(map[string]orders.CabOrders, len(m))
		for id, cab := range m {
			if cab == nil {
				continue
			}
			cc := cab.Copy()
			cp[id] = *cc
		}
		return cp, true
	}
	copyElevStateValue := func(m map[string]*elevator.Elevator) (map[string]elevator.Elevator, bool) {
		if m == nil {
			return nil, false
		}
		cp := make(map[string]elevator.Elevator, len(m))
		for id, e := range m {
			if e == nil {
				continue
			}
			cp[id] = *e
		}
		return cp, true
	}

	go func() {
		var (
			currentMergedHall orders.HallOrders
			currentAllCab     map[string]orders.CabOrders
			currentElevState  map[string]elevator.Elevator
		)

		publish := func() {
			myCab, ok := currentAllCab[config.MyID]
			if !ok {
				myCab = orders.CabOrders{}
			}

			chMsg := CallHandlerMessage{
				mergedHallOrders: currentMergedHall,
				myCabOrders:      myCab,
			}
			odMsg := OrderDistributorMessage{
				mergedHallOrders: currentMergedHall,
				allCabOrders:     currentAllCab,
				elevatorState:    currentElevState,
			}
			netMsg := NetworkingDistributorMessage{
				allCabOrders:     currentAllCab,
				mergedHallOrders: currentMergedHall,
				elevatorState:    currentElevState,
			}

			//Start by emptying all the channels
			select {
			case <-callHandlerOutput:
			default:
			}

			select {
			case <-orderDistributorOutput:
			default:
			}

			select {
			case <-networkingDistributorOutput:
			default:
			}
			// Then writing your new message to the channels
			channelForCallHandler <- chMsg
			channelForOrderDistributor <- odMsg
			channelForNetworking <- netMsg
		}

		for {
			select {
			case h := <-hallOut:
				if hv, ok := copyHallValue(h); ok {
					currentMergedHall = hv
				}
				publish()

			case c := <-cabOut:
				if cv, ok := copyAllCabValue(c); ok {
					currentAllCab = cv
				}
				publish()

			case es := <-elevatorState:
				if ev, ok := copyElevStateValue(es); ok {
					currentElevState = ev
				}
				publish()
			}
		}
	}()
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
	channelToCallHandler <-chan CallHandlerMessage,
	channelToOrderDistributor <-chan OrderDistributorMessage,
	channelToNetworking chan <- NetworkingDistributorMessage,
	channelFromNetworking <- chan NetworkingDistributorMessage,
) {
	allHall := map[string]*orders.HallOrders{}
	allCab := map[string]*orders.CabOrders{}
	allElevatorStates := map[string]*elevator.Elevator{}
	elevatorsOnNetwork := []string{}

	// Create your own orders first
	allHall[config.MyID] = orders.CreateHallOrders()
	allCab[config.MyID] = orders.CreateCabOrders()
	initialElevatorState := <-elevatorStateUpdate
	allElevatorStates[config.MyID] = &initialElevatorState

	hallOut := make(chan *orders.HallOrders, 1)
	cabOut := make(chan map[string]*orders.CabOrders, 1)

	// Internal snapshot channels: update loop sends latest state, broadcast goroutine reads it.
	hallSnap := make(chan *orders.HallOrders, 1)
	cabSnap := make(chan map[string]*orders.CabOrders, 1)


	// Må tenkte litt mer, men mulig vi kan fjerne. Har ikke kontinuerlig oppdatering av variabler.
	go func() {
		ticker := time.NewTicker(config.HeartbeatInterval)
		defer ticker.Stop()
		latestHall := allHall[config.MyID]
		latestCab := orders.CopyAllCab(allCab)
		for {
			select {
			case h := <-hallSnap:
				latestHall = h
			case c := <-cabSnap:
				latestCab = c
			case <-ticker.C:
				hallOut <- latestHall
				cabOut <- latestCab
			}
		}
	}()

	go func() {
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
			elevCopy := tempElevator[message.SenderID()].Copy()
			elevatorStateUpdate <- elevCopy
		}
	}()

	for {
		select {
		case u := <-hallUpdate:
			if _, ok := allHall[u.SenderID]; !ok {
				// Burde kanskje sjekke at u.SenderID faktisk er en elevator. Mulig det skjer i Networking.
				allHall[u.SenderID] = orders.CreateHallOrders()
			}
			allHall[u.SenderID].UpdateOrderState(u.Floor, u.Direction, u.State)
			nextState := mergeHallOrderState(u, config.MyID, allHall, elevatorsOnNetwork)
			allHall[config.MyID].UpdateOrderState(u.Floor, u.Direction, nextState)
			// Empty the channel to always have the lates snapshot
			select {
			case <-hallSnap:
			default:
			}
			hallSnap <- allHall[config.MyID].Copy()

		case u := <-cabUpdate:
			if _, ok := allCab[u.SenderID]; !ok {
				// Burde kanskje sjekke at u.SenderID faktisk er en elevator. Mulig det skjer i Networking.
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
			// Always overwrite the elevator state for the sender, since it's a direct report of its physical state, not a distributed consensus like the orders.
			allElevatorStates[es.Id()] = &es
		}
	}
}