package messagestruct

import  (
	es "elevatorproject/internal/elevatorstruct"
)

type Message struct {
	elevators map[string]es.Elevator
}

// Adds a new elevator to the map of elevators in the message struct. 
// If the elevator already exists, State Machine!!

// handleElevatorStates()
// handleOrderStates(){
//     handleCabOrders()
//     handleHallOrders()
// }



// t = 0
// elev1.HallRequests = [unknown, unconfirmed]
// elev2.HallRequests = [unknown, unknown]

// t = 1
// elev1.HallRequests = [unknown, unconfirmed]
// elev2.HallRequests = [unknown, unconfirmed]

// t = 2
// elev1.HallRequests = [unknown, confirmed]
// elev2.HallRequests = [unknown, confirmed]

