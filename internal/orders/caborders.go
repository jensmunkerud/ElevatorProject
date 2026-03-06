package orders


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

