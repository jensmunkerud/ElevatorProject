package orders

import "fmt"

// All possible states of an order, Unconfirmed and finished act as barriers.
type OrderState int

const (
	UnknownOrderState OrderState = iota
	RemovedOrderState
	UnconfirmedOrderState
	ConfirmedOrderState
	CompletedOrderState
)

type Order struct {
	state OrderState
}

func CreateOrder() *Order {
	return &Order{
		state: UnknownOrderState,
	}
}

func (o *Order) UpdateState(state OrderState) {
	if !state.IsValid() {
		fmt.Printf("orders: invalid order state %d\n", state)
		return
	}
	o.state = state
}

func (o *Order) GetState() OrderState {
	return o.state
}

func (o OrderState) IsUnknown() bool {
	return o == UnknownOrderState
}

func (o OrderState) IsCompleted() bool {
	return o == CompletedOrderState
}

func (o OrderState) IsRemoved() bool {
	return o == RemovedOrderState
}

func (o OrderState) IsUnconfirmed() bool {
	return o == UnconfirmedOrderState
}

func (o OrderState) IsConfirmed() bool {
	return o == ConfirmedOrderState
}

func (o OrderState) IsValid() bool {
	return o >= UnknownOrderState && o <= CompletedOrderState
}
