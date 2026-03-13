package orders

import "elevatorproject/src/config"

type HallOrders struct {
	Orders [][2]*Order //[floor][up/down] for each floor and direction
}

// A [NumFloors][2] array of pointers to OrderState,
// where the first index represents the floor
// and the second index represents the direction (0 for up, 1 for down).
func CreateHallOrders() *HallOrders {
	orders := make([][2]*Order, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		up := CreateOrder()
		down := CreateOrder()
		orders[i][0] = up
		orders[i][1] = down
	}
	return &HallOrders{Orders: orders}
}

func (o *HallOrders) HallOrderState(floor int, direction int) OrderState {
	return o.Orders[floor][direction].GetState()
}

// Simplify converts HallOrders to a simpler [][]bool format for easier processing in cost functions.
// Returns true for states Confirmed and Completed, false otherwise.
func (h *HallOrders) Simplify() [][]bool {
	simplified := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		simplified[floor] = make([]bool, 2)
		for direction := 0; direction < 2; direction++ {
			simplified[floor][direction] = h.HallOrderState(floor, direction) == ConfirmedOrderState ||
				h.HallOrderState(floor, direction) == CompletedOrderState
		}
	}
	return simplified
}

func (h *HallOrders) UpdateOrderState(floor int, direction int, state OrderState) {
	h.Orders[floor][direction].UpdateState(state)
}

func (h *HallOrders) GetOrderState(floor int, direction int) OrderState {
	return h.Orders[floor][direction].GetState()
}

func (h *HallOrders) Copy() *HallOrders {
	cp := CreateHallOrders()
	for floor := 0; floor < config.NumFloors; floor++ {
		for dir := 0; dir < 2; dir++ {
			cp.Orders[floor][dir].UpdateState(h.Orders[floor][dir].GetState())
		}
	}
	return cp
}