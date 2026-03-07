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
