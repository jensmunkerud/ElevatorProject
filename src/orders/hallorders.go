package orders

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"fmt"
)

type HallOrders struct {
	orders [][2]*Order //[floor][up/down] for each floor and orderType
}

// A [NumFloors][2] array of pointers to OrderState,
// where the first index represents the floor
// and the second index represents the orderType (0 for up, 1 for down).
func CreateHallOrders() *HallOrders {
	orders := make([][2]*Order, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		orders[floor][0] = CreateOrder()
		orders[floor][1] = CreateOrder()
	}
	return &HallOrders{orders: orders}
}

func (o *HallOrders) HallOrderState(floor int, orderType elevator.OrderType) OrderState {
	if !isValidFloor(floor) || !isValidHallOrderType(orderType) {
		return UnknownOrderState
	}
	return o.orders[floor][int(orderType)].GetState()
}

// Simplify converts HallOrders to a simpler [][]bool format for easier processing in cost functions.
// Returns true for states Confirmed, otherwise false
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
	if !isValidFloor(floor) || !isValidHallOrderType(orderType) {
		return
	}
	h.orders[floor][int(orderType)].UpdateState(state)
}

func (h *HallOrders) GetOrderState(floor int, orderType elevator.OrderType) OrderState {
	if !isValidFloor(floor) || !isValidHallOrderType(orderType) {
		return UnknownOrderState
	}
	return h.orders[floor][int(orderType)].GetState()
}

func (h *HallOrders) Copy() *HallOrders {
	copy := CreateHallOrders()
	for floor := 0; floor < config.NumFloors; floor++ {
		for _, orderType := range elevator.HallOrderTypes {
			copy.orders[floor][int(orderType)].UpdateState(h.orders[floor][int(orderType)].GetState())
		}
	}
	return copy
}

func isValidHallOrderType(orderType elevator.OrderType) bool {
	if orderType != elevator.HallUp && orderType != elevator.HallDown {
		fmt.Printf("orders: invalid hall order type %d\n", orderType)
		return false
	}
	return true
}
