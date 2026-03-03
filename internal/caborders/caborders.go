package caborders

import (
	orderstruct "elevatorproject/internal/orderStruct"
)

type CabOrders struct {
	Orders []*orderstruct.OrderState
}

//A [NumFloors][2] array of pointers to OrderState, 
// where the first index represents the floor 
// and the second index represents the direction (0 for up, 1 for down).
func CreateCabOrders(numFloors int) *CabOrders {
	orders := make([]*orderstruct.OrderState, numFloors)
	for i := 0; i < numFloors; i++ {
		currentOrder := orderstruct.OrderStateUnknown
		orders[i] = &currentOrder
	}
	return &CabOrders{Orders: orders}
}


func (o *CabOrders) CabOrderState(floor int) orderstruct.OrderState {
	return *o.Orders[floor]
}

