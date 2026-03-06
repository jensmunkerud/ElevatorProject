package orders

import "elevatorproject/internal/config"

//Indexed by floor, contains the state of the cab order for each floor.
type CabOrders struct {
	Orders []*OrderState
}

//A [NumFloors][2] array of pointers to OrderState, 
// where the first index represents the floor 
// and the second index represents the direction (0 for up, 1 for down).
func CreateCabOrders(numFloors int) *CabOrders {
	orders := make([]*OrderState, numFloors)
	for i := 0; i < numFloors; i++ {
		currentOrder := OrderStateUnknown
		orders[i] = &currentOrder
	}
	return &CabOrders{Orders: orders}
}


func (o *CabOrders) CabOrderState(floor int) OrderState {
	return *o.Orders[floor]
}

// Simplify converts CabOrders to a simpler []bool format for easier processing in cost functions.
// Returns true for states Confirmed and Completed, false otherwise.
func (c *CabOrders) Simplify() []bool {
	simplified := make([]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		simplified[floor] = c.CabOrderState(floor) == OrderStateConfirmed ||
			c.CabOrderState(floor) == OrderStateCompleted
	}
	return simplified
}