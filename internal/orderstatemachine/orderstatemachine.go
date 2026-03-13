package orderstatemachine

import (
	orders "elevatorproject/internal/orders"
)

type OrderStateMachine struct {
	currentState orders.OrderState
}

type OrderStruct struct{
	floor int 
	direction string
	cabOrder bool
	hallOrder bool
	confirmedBy string
	completed bool
	elevID UUID
}

// [[stateUnconfirmed, confirmedState][unknownState, unknownState]]
// -> floor
// -> direction
// -> CabOrders or HallOrders
// -> Nei, ikke confirmedBy
// -> Ja, sjekk om orderComplete, men ikke av hvem
// -> Nei, ikke hvem som har sendt orderen

type activeOrders struct {
	activeCabOrders []OrderStruct
	activeHallOrders []OrderStruct
}

// Creates a new OrderStateMachine and initializes it to the
// Unknown state. The OrderStateMachine is sent to the channel
// every time it is updated, so that other parts of the program can
// react to changes in the state of the elevator.
func CreateOrderStateMachine(
	// Messagehandler needs to create a channel that checks if
	// an order is confirmed.
) chan OrderStateMachine {
	orderStateMachine := make(chan OrderStateMachine)

	go func() {
		// By default, an OrderStateMachine should be put in Unknown
		// state, since we don't know the state of the elevator at startup
		previousState := orders.OrderStateUnknown
		orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnknown}
		for {
			select {
			case isOrderConfirmed := <-*isOrderConfirmed:
				if isOrderConfirmed && previousState == orders.OrderStateUnconfirmed {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateConfirmed}
					previousState = orders.OrderStateConfirmed
				} else {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnknown}
					previousState = orders.OrderStateUnknown
				}
			case hasCabOrder := <-*hasNewCabOrder:
				if previousState == orders.OrderStateUnknown && hasCabOrder {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnconfirmed}
					previousState = orders.OrderStateUnconfirmed
				} else if hasCabOrder && previousState == orders.OrderStateUnconfirmed {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateConfirmed}
					previousState = orders.OrderStateConfirmed
				} else {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnknown}
				}
			case hasHallOrder := <-*hasNewHallOrder:
				if previousState == orders.OrderStateUnknown && hasHallOrder {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnconfirmed}
					previousState = orders.OrderStateUnconfirmed
				} else if hasHallOrder && previousState == orders.OrderStateUnconfirmed {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateConfirmed}
					previousState = orders.OrderStateConfirmed
				} else {
					orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnknown}
				}
			default:
				orderStateMachine <- OrderStateMachine{currentState: orders.OrderStateUnknown}
			}
		}
	}()
	return orderStateMachine
}


func checkIfOrderIsConfirmed(ordre) chan bool {

func checkIfHasNewCabOrder(ordre) chan bool {

func checkIfHasNewHallOrder(ordre) chan bool {


func checkIfHasCompletedOrder(ordre) chan bool {