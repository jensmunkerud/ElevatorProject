package orders

// All possible states of an order, Unconfirmed and finished act as barriers.
type OrderState int

const (
	UnknownOrderState OrderState = iota
	RemovedOrderState
	UnconfirmedOrderState
	ConfirmedOrderState
	CompletedOrderState
)

type Order struct{
	State OrderState
}
	
func CreateOrder() *Order {
	return &Order{
		State: UnknownOrderState,
	}
}

func (o *Order) UpdateState(state OrderState) {
	o.State = state
}

func (o *Order) GetState() OrderState {
	return o.State
}

func (o OrderState) Unknown() bool {
	return o == UnknownOrderState
}

func (o OrderState) Completed() bool {
	return o == CompletedOrderState
}

func (o OrderState) Removed() bool {
	return o == RemovedOrderState
}

func (o OrderState) Unconfirmed() bool {
	return o == UnconfirmedOrderState
}

func (o OrderState) Confirmed() bool {	
	return o == ConfirmedOrderState
}