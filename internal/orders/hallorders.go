package orders


type HallOrders struct {
	Orders [][2]*OrderState //[floor][up/down] for each floor and direction
}



//A [NumFloors][2] array of pointers to OrderState, 
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

