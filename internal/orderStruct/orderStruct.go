package orderstruct

type OrderState int

const (
	OrderStateUnknown OrderState = iota
	OrderStateNone
	OrderStateUnconfirmed
	OrderStateConfirmed
)
