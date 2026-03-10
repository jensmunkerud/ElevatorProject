package messagestruct

import  (
	es "elevatorproject/src/elevatorStruct"
	os "elevatorproject/src/orderStruct"
)

type Message struct {
	elev1 es.Elevator
	elev2 es.Elevator
	elev3 es.Elevator

	orders os.Orders
}