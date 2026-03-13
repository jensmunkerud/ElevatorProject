package orders

import "elevatorproject/src/config"

//Indexed by floor, contains the state of the cab order for each floor.
type CabOrders struct {
	Orders []*Order
}

//A [NumFloors][2] array of pointers to OrderState,
// where the first index represents the floor
// and the second index represents the direction (0 for up, 1 for down).
func CreateCabOrders() *CabOrders {
	orders := make([]*Order, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		order := CreateOrder()
		orders[i] = order
	}
	return &CabOrders{Orders: orders}
}

func (o *CabOrders) CabOrderState(floor int) OrderState {
	return o.Orders[floor].GetState()
}

// Simplify converts CabOrders to a simpler []bool format for easier processing in cost functions.
// Returns true for states Confirmed and Completed, false otherwise.
func (c *CabOrders) Simplify() []bool {
	simplified := make([]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		simplified[floor] = c.CabOrderState(floor) == ConfirmedOrderState ||
			c.CabOrderState(floor) == CompletedOrderState
	}
	return simplified
}

func (c *CabOrders) UpdateOrderState(floor int, state OrderState) {
	c.Orders[floor].UpdateState(state)
}

func (c *CabOrders) GetOrderState(floor int) OrderState {
	return c.Orders[floor].GetState()
}

func (c *CabOrders) Copy() *CabOrders {
	cp := CreateCabOrders()
	for floor := 0; floor < config.NumFloors; floor++ {
		cp.Orders[floor].UpdateState(c.Orders[floor].GetState())
	}
	return cp
}

// CopyAllCab returns a deep copy of a map[string]*CabOrders.
func CopyAllCab(m map[string]*CabOrders) map[string]*CabOrders {
	cp := make(map[string]*CabOrders, len(m))
	for id, cab := range m {
		cp[id] = cab.Copy()
	}
	return cp
}