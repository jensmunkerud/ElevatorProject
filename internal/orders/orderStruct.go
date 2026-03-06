package orders

//All possible states of an order, Unconfirmed and finished act as barriers.
type OrderState int

const (
	OrderStateUnknown OrderState = iota
	OrderStateNone
	OrderStateUnconfirmed
	OrderStateConfirmed
	OrderStateCompleted
)
