package networking

import  (
	es "elevatorproject/src/elevator"
	os "elevatorproject/src/orders"
)

type Message struct {
	elev1 es.Elevator
	elev2 es.Elevator
	elev3 es.Elevator

	orders os.Orders
}




/*

På heis 1
if Order1Heis1 == Order1Heis2 == Order1Heis3 == ElevatorState.Unconfirmed {
	OrderHeis1 = ElevatorState.Confirmed
}

På heis 2
if Order1Heis1 == Order1Heis2 == Order1Heis3 == ElevatorState.Unconfirmed {
	OrderHeis2 = ElevatorState.Confirmed
}

På heis 3
if Order1Heis1 == Order1Heis2 == Order1Heis3 == ElevatorState.Unconfirmed {
	OrderHeis3 = ElevatorState.Confirmed
}



*/