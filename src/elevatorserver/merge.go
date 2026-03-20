package elevatorserver

import (
	"elevatorproject/src/orders"
	"fmt"
)

//This file contains the logic for merging incoming order updates from the network with the local
// state of orders, using a distributed barrier protocol to ensure consistency across all online nodes.
// This allows the system to preserve orders even in the face of elevator failures and recoveries.

// mergeHallOrderState resolves the next OrderState for a hall order at the given floor and direction.
// It compares the incoming state from update.SenderID against the receiver's local state, using the
// barrier protocol to coordinate transitions across all online nodes. This preserves orders across elevator failures.
func mergeHallOrderState(update HallOrderUpdate, receiverID string, allOrders map[string]*orders.HallOrders, onlineNodes []string) orders.OrderState {
	local := allOrders[receiverID].GetOrderState(update.Floor, update.OrderType)
	noOtherOnlineNodes := true
	for _, id := range onlineNodes {
		if id != receiverID {
			noOtherOnlineNodes = false
		}
	}
	removed := local.Removed()
	unknown := local.Unknown()
	if (removed || unknown) && noOtherOnlineNodes {
		return orders.UnknownOrderState
	}
	return mergeState(update.State, local, onlineNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allOrders[id]
		if !ok {
			return orders.UnknownOrderState, false
		}
		return elev.GetOrderState(update.Floor, update.OrderType), true
	})
}

// Merges an incoming cab order state with the local state, using the same barrier protocol as mergeHallOrderState to coordinate transitions across all online nodes.
// Cab orders are per-elevator, so the barrier only checks the owning
// elevator's own state (not other elevators' unrelated cab orders).
func mergeCabOrderState(update CabOrderUpdate, allCabOrders map[string]*orders.CabOrders, onlineNodes []string) orders.OrderState {
	local := allCabOrders[update.SenderID].GetOrderState(update.Floor)

	// Once a cab order has been cleared (Removed), don't allow stale
	// Confirmed from peers to resurrect it. Only Unconfirmed (a new button
	// press) restarts the cycle. Without this, a peer's stale heartbeat
	// containing Confirmed overwrites Removed via "highest state wins"
	// because Confirmed(3) > Removed(1).
	//if local == orders.RemovedOrderState && update.State == orders.ConfirmedOrderState {
	//	return orders.RemovedOrderState
	//}

	// Only the owning elevator participates in the barrier for cab orders.
	fmt.Printf("New cab order update: %v\n", update)
	noOtherOnlineNodes := true
	cabBarrierNodes := []string{}
	for _, id := range onlineNodes {
		if id != update.SenderID {
			noOtherOnlineNodes = false
		}
	}
	if noOtherOnlineNodes {
		cabBarrierNodes = []string{update.SenderID}
	} else {
		cabBarrierNodes = onlineNodes
	}
	return mergeState(update.State, local, cabBarrierNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allCabOrders[id]
		if !ok {
			return orders.UnknownOrderState, false
		}
		return elev.GetOrderState(update.Floor), true
	})
}

// mergeState contains the shared state-machine logic for merging an incoming order state with
// the local state. It applies two distributed barriers: Unconfirmed→Confirmed and Completed→Removed (all nodes must
// have seen the order).
// getState is a callback that retrieves a node's current OrderState by ID, returning false if the node is unknown.
func mergeState(newOrder orders.OrderState, local orders.OrderState, onlineNodes []string, getState func(string) (orders.OrderState, bool)) orders.OrderState {
	fmt.Printf("Merging state: local: %v, newOrder: %v\n", local, newOrder)
	fmt.Printf("Online nodes: %v\n", onlineNodes)
	switch local {
	case orders.UnknownOrderState:
		return newOrder
	case orders.RemovedOrderState:
		if newOrder.Unconfirmed() {
			// Need to check barrier for single elevator case
			if barrierReached(onlineNodes, orders.UnconfirmedOrderState, getState) {
				fmt.Printf("Merge of 2 and 2 resulted in confirmed \n")
				return orders.ConfirmedOrderState
			} else {
				fmt.Printf("Merge of 2 and 2 resulted in unconfirmed \n")
				return orders.UnconfirmedOrderState
			}
		} else {
			return local
		}
	case orders.UnconfirmedOrderState:
		if barrierReached(onlineNodes, orders.UnconfirmedOrderState, getState) {
			return orders.ConfirmedOrderState
		} else {
			return local
		}
	case orders.ConfirmedOrderState:
		if newOrder.Completed() {
			// Need to check barrier for single elevator case
			if barrierReached(onlineNodes, orders.CompletedOrderState, getState) {
				return orders.RemovedOrderState
			} else {
				return orders.CompletedOrderState
			}
		} else {
			return local
		}
	case orders.CompletedOrderState:
		if barrierReached(onlineNodes, orders.CompletedOrderState, getState) {
			return orders.RemovedOrderState
		} else {
			return local
		}
	default:
		return local
	}
}

// barrierReached returns true if every online node has reached the given threshold.
// For the Completed threshold, a node at Removed is treated as having passed through
// Completed (the lifecycle is Unconfirmed→Confirmed→Completed→Removed, but Removed
// has a lower numeric value than Completed).
func barrierReached(onlineNodes []string, threshold orders.OrderState, getState func(string) (orders.OrderState, bool)) bool {
	for _, id := range onlineNodes {
		state, ok := getState(id)
		if !ok {
			return false
		}
		if threshold == orders.CompletedOrderState {
			if state != orders.CompletedOrderState && state != orders.RemovedOrderState {
				return false
			}
		} else if state < threshold {
			return false
		}
	}
	return true
}
