package messagestruct

import  (
	es "elevatorproject/internal/elevatorStruct"
	os "elevatorproject/internal/orderStruct"
)

type Message struct {
	elev1 es.Elevator
	elev2 es.Elevator
	elev3 es.Elevator

	orders os.Orders
}