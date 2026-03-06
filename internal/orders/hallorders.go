package orders

import "elevatorproject/internal/config"

type HallOrders struct {
	Orders [][2]*OrderState //[floor][up/down] for each floor and direction
}

// A [NumFloors][2] array of pointers to OrderState,
// where the first index represents the floor
// and the second index represents the direction (0 for up, 1 for down).
func CreateHallOrders(numFloors int) *HallOrders {
	orders := make([][2]*OrderState, numFloors)
	for i := 0; i < numFloors; i++ {
		up := OrderStateUnknown
		down := OrderStateUnknown
		orders[i][0] = &up
		orders[i][1] = &down
	}
	return &HallOrders{Orders: orders}
}

func (o *HallOrders) HallOrderState(floor int, direction int) OrderState {
	return *o.Orders[floor][direction]
}

// Simplify converts HallOrders to a simpler [][]bool format for easier processing in cost functions.
// Returns true for states Confirmed and Completed, false otherwise.
func (h *HallOrders) Simplify() [][]bool {
	simplified := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		simplified[floor] = make([]bool, 2)
		for direction := 0; direction < 2; direction++ {
			simplified[floor][direction] = h.HallOrderState(floor, direction) == OrderStateConfirmed ||
				h.HallOrderState(floor, direction) == OrderStateCompleted
		}
	}
	return simplified
}
