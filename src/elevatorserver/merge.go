package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/orders"
	"fmt"
)

//This file contains the logic for merging incoming order updates from the network with the local
// state of orders, using a distributed barrier protocol to ensure consistency across all online nodes.
// This allows the system to preserve orders even in the face of elevator failures and recoveries.

// Checks if all elevators have seen an order before confirming or removing it.
func mergeHallOrderState(update HallOrderUpdate, receiverID string, allOrders map[string]*orders.HallOrders, onlineNodes []string) orders.OrderState {
	local := allOrders[receiverID].GetOrderState(update.Floor, update.OrderType)
	noOtherOnlineNodes := true
	for _, id := range onlineNodes {
		if id != receiverID {
			noOtherOnlineNodes = false
		}
	}
	removed := local.IsRemoved()
	unknown := local.IsUnknown()
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

// Checks if any other elevator has seen our cab order before confirming and removing the order.
func mergeCabOrderState(update CabOrderUpdate, allCabOrders map[string]*orders.CabOrders, onlineNodes []string) orders.OrderState {
	local := allCabOrders[update.OwnerID].GetOrderState(update.Floor)
	myID := config.MyID()
	hasOtherOnlineNodes := false
	for _, id := range onlineNodes {
		if id != myID {
			hasOtherOnlineNodes = true
			break
		}
	}

	// Only transition if a peer has confirmed receival of the order.
	if update.OwnerID == myID && local == orders.UnconfirmedOrderState {
		if !hasOtherOnlineNodes {
			return orders.ConfirmedOrderState
		}
		if update.SenderID != myID && update.State >= orders.UnconfirmedOrderState {
			return orders.ConfirmedOrderState
		}
		return orders.UnconfirmedOrderState
	}

	if update.SenderID == myID &&
		update.OwnerID != myID &&
		local == orders.UnconfirmedOrderState &&
		update.State >= orders.UnconfirmedOrderState {
		return orders.ConfirmedOrderState
	}

	cabBarrierNodes := []string{update.OwnerID}
	return mergeState(update.State, local, cabBarrierNodes, func(id string) (orders.OrderState, bool) {
		elev, ok := allCabOrders[id]
		if !ok {
			return orders.UnknownOrderState, false
		}
		return elev.GetOrderState(update.Floor), true
	})
}

// mergeState contains the shared state-machine logic for merging an incoming order state with
// the local state.
func mergeState(newOrder orders.OrderState, local orders.OrderState, onlineNodes []string, getState func(string) (orders.OrderState, bool)) orders.OrderState {
	switch local {
	case orders.UnknownOrderState:
		return newOrder
	case orders.RemovedOrderState:
		if newOrder.IsUnconfirmed() {
			if barrierReached(onlineNodes, orders.UnconfirmedOrderState, getState) {
				return orders.ConfirmedOrderState
			} else {
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
		if newOrder.IsCompleted() {
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
func barrierReached(onlineNodes []string, threshold orders.OrderState, getState func(string) (orders.OrderState, bool)) bool {

	for _, id := range onlineNodes {
		state, ok := getState(id)
		if !ok {
			fmt.Printf("Node %v not found\n", id)
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
