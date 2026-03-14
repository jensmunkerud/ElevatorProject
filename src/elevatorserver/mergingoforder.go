package elevatorserver

import (
	"elevatorproject/src/orders"
)

// mergingOfOrders should be used to keep track
// of the state of each order, found in ./Documentation/orderState
// The states are unknown, unconfirmed, confirmed, completed and none
// Unknown is the initial-state of an order, and it can transition to any other state from it.
// After this, the order can stricty move from none -> unconfirmed -> confirmed -> completed.


// Function that takes in three (should be expanded to numOfAliveElevators) caborders and returns a merged caborders for further use
func mergeCabOrders(allIncomingCabOrders []orders.CabOrders,
	mergedCabOrders *orders.CabOrders) {
	for _, order := range allIncomingCabOrders[0].Orders {
		switch currentState := order.GetState(); currentState {
		case orders.UnknownOrderState:
			//mergeUnknownOrders()
		case orders.UnconfirmedOrderState:
			//mergeUnconfirmedOrders()
		case orders.ConfirmedOrderState:
		case orders.CompletedOrderState:
		case orders.RemovedOrderState:
		}
	}
}

func addCabOrder(newOrder orders.Order, mergedCabOrders *orders.CabOrders) {
	// Check if newOrder is valid to add in mergedCabOrders, 
	// if so, add it and update the state of the order in mergedCabOrders
}

//Function that takes in three (should be expanded to numOfAliveElevators) hallorders and returns a merged hallorders for further use

