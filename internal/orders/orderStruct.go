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