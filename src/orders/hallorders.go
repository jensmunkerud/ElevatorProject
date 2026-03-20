package orders

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
)

type HallOrders struct {
	Orders [][2]*Order //[floor][up/down] for each floor and orderType
}

// A [NumFloors][2] array of pointers to OrderState,
// where the first index represents the floor
// and the second index represents the orderType (0 for up, 1 for down).
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

func (o *HallOrders) HallOrderState(floor int, orderType elevator.OrderType) OrderState {
	return o.Orders[floor][int(orderType)].GetState()
}

// Simplify converts HallOrders to a simpler [][]bool format for easier processing in cost functions.
// Returns true for states Confirmed and Completed, false otherwise.
func (h *HallOrders) Simplify() [][]bool {
	simplified := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		simplified[floor] = make([]bool, 2)
		for _, orderType := range elevator.HallOrderTypes {
			simplified[floor][int(orderType)] = h.HallOrderState(floor, orderType) == ConfirmedOrderState
		}
	}
	return simplified
}

func (h *HallOrders) UpdateOrderState(floor int, orderType elevator.OrderType, state OrderState) {
	h.Orders[floor][int(orderType)].UpdateState(state)
}

func (h *HallOrders) GetOrderState(floor int, orderType elevator.OrderType) OrderState {
	return h.Orders[floor][int(orderType)].GetState()
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
