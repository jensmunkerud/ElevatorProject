package orders

import (
	"elevatorproject/src/config"
	"fmt"
)

// Indexed by floor, contains the state of the cab order for each floor.
type CabOrders struct {
	orders []*Order
}

// A [NumFloors] array of pointers to OrderState,
func CreateCabOrders() *CabOrders {
	orders := make([]*Order, config.NumFloors)
	for floor := range config.NumFloors {
		order := CreateOrder()
		orders[floor] = order
	}
	return &CabOrders{orders: orders}
}

func (o *CabOrders) CabOrderState(floor int) OrderState {
	if !isValidFloor(floor) {
		return UnknownOrderState
	}
	return o.orders[floor].GetState()
}

// Simplify converts CabOrders to a simpler []bool format for easier processing in cost functions.
// Returns true for states Confirmed, otherwise false.
func (c *CabOrders) Simplify() []bool {
	simplified := make([]bool, config.NumFloors)
	for floor := range config.NumFloors {
		simplified[floor] = c.CabOrderState(floor) == ConfirmedOrderState
	}
	return simplified
}

func (c *CabOrders) UpdateOrderState(floor int, state OrderState) {
	if !isValidFloor(floor) {
		return
	}
	c.orders[floor].UpdateState(state)
}

func (c *CabOrders) GetOrderState(floor int) OrderState {
	if !isValidFloor(floor) {
		return UnknownOrderState
	}
	return c.orders[floor].GetState()
}

func isValidFloor(floor int) bool {
	if floor < 0 || floor >= config.NumFloors {
		fmt.Printf("orders: invalid floor %d (valid range 0-%d)\n", floor, config.NumFloors-1)
		return false
	}
	return true
}

func (c *CabOrders) Copy() *CabOrders {
	copy := CreateCabOrders()
	for floor := range config.NumFloors {
		copy.orders[floor].UpdateState(c.orders[floor].GetState())
	}
	return copy
}

func CopyAllCab(cabOrders map[string]*CabOrders) map[string]*CabOrders {
	copy := make(map[string]*CabOrders, len(cabOrders))
	for id, cab := range cabOrders {
		copy[id] = cab.Copy()
	}
	return copy
}
