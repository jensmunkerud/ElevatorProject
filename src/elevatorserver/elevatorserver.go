package elevatorserver

import (
	"elevatorproject/src/config"
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

//This should be moved to networking.
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

//This should be moved to networking.
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
// It compares the incoming state from update.SenderID against the receiver's local state, using the
// barrier protocol to coordinate transitions across all online nodes.
func mergeCabOrderState(update CabOrderUpdate, receiverID string, allOrders map[string]*orders.CabOrders, onlineNodes []string) orders.OrderState {
	local := allOrders[receiverID].GetOrderState(update.Floor)
	return mergeState(update.State, local, onlineNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allOrders[id]
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

// RunElevatorServer runs the elevator server.
// hallOut carries the receiver's updated hall orders at a fixed broadcast interval.
// cabOut carries a snapshot of all known elevators' cab orders each interval, so peers
// can restore orders for an elevator that has gone offline.
func RunElevatorServer(
	hallUpdates <-chan HallOrderUpdate,
	cabUpdates <-chan CabOrderUpdate,
	peerUpdates <-chan []string,
	hallOut chan<- *orders.HallOrders,
	cabOut chan<- map[string]*orders.CabOrders,
	receiverID string,
) {
	allHall := map[string]*orders.HallOrders{}
	allCab := map[string]*orders.CabOrders{}
	online := []string{}

	allHall[receiverID] = orders.CreateHallOrders()
	allCab[receiverID] = orders.CreateCabOrders()

	// Internal snapshot channels: update loop sends latest state, broadcast goroutine reads it.
	hallSnap := make(chan *orders.HallOrders, 1)
	cabSnap := make(chan map[string]*orders.CabOrders, 1)

	go func() {
		ticker := time.NewTicker(config.HeartbeatInterval)
		defer ticker.Stop()
		latestHall := allHall[receiverID]
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

	for {
		select {
		case u := <-hallUpdates:
			if _, ok := allHall[u.SenderID]; !ok {
				allHall[u.SenderID] = orders.CreateHallOrders()
			}
			allHall[u.SenderID].UpdateOrderState(u.Floor, u.Direction, u.State)
			next := mergeHallOrderState(u, receiverID, allHall, online)
			allHall[receiverID].UpdateOrderState(u.Floor, u.Direction, next)
			select {
			case <-hallSnap:
			default:
			}
			hallSnap <- allHall[receiverID].Copy()

		case u := <-cabUpdates:
			if _, ok := allCab[u.SenderID]; !ok {
				allCab[u.SenderID] = orders.CreateCabOrders()
			}
			allCab[u.SenderID].UpdateOrderState(u.Floor, u.State)
			next := mergeCabOrderState(u, receiverID, allCab, online)
			allCab[receiverID].UpdateOrderState(u.Floor, next)
			select {
			case <-cabSnap:
			default:
			}
			cabSnap <- orders.CopyAllCab(allCab)

		case nodes := <-peerUpdates:
			online = nodes
			for _, id := range nodes {
				if _, ok := allHall[id]; !ok {
					allHall[id] = orders.CreateHallOrders()
				}
				if _, ok := allCab[id]; !ok {
					allCab[id] = orders.CreateCabOrders()
				}
			}
		}
	}
}
