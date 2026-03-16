package elevatorserver

import (
	"elevatorproject/src/orders"
)

// mergeHallOrderState resolves the next OrderState for a hall order at the given floor and direction.
// It compares the incoming state from update.SenderID against the receiver's local state, using the
// barrier protocol to coordinate transitions across all online nodes. This preserves orders across elevator failures.
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

// Merges an incoming cab order state with the local state, using the same barrier protocol as mergeHallOrderState to coordinate transitions across all online nodes.
// This ensures that cab orders are preserved even if an elevator goes offline and later comes back online.
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
// the local state. It applies two distributed barriers: Unconfirmed→Confirmed and Completed→Removed (all nodes must
// have seen the order).
// getState is a callback that retrieves a node's current OrderState by ID, returning false if the node is unknown.
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
