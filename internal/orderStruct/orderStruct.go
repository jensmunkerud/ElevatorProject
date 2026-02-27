package orderstruct

type OrderState int

const (
	OrderStateUnknown OrderState = iota
	OrderStateNone
	OrderStateUnconfirmed
	OrderStateConfirmed
)

type Order struct {
	State   OrderState
	AckedBy map[string]bool
}

type Orders struct {
	Orders [][2]Order //[floor][up/down] for each floor and direction
}

func (o *Orders) Initialize(numFloors int) {
	o.Orders = make([][2]Order, numFloors)
	for i := 0; i < numFloors; i++ {
		o.Orders[i][0] = Order{State: OrderStateUnknown, AckedBy: make(map[string]bool)}
		o.Orders[i][1] = Order{State: OrderStateUnknown, AckedBy: make(map[string]bool)}
	}
}

func (o *Orders) OrderState(floor int, direction int) OrderState {
	return o.Orders[floor][direction].State
}

// AllAckedCheck returns true if all elevators in requiredIDs have acknowledged the order
func (o *Orders) AllAckedCheck(floor int, direction int, requiredIDs []string) bool {
	ackedBy := o.Orders[floor][direction].AckedBy
	for _, id := range requiredIDs {
		if !ackedBy[id] {
			return false
		}
	}
	return true
}

func (o *Orders) SetOrderState(floor int, direction int, state OrderState) {
	o.Orders[floor][direction].State = state
}