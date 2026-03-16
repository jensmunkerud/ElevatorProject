package orderdistributor

import (
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"encoding/json"
	"fmt"
)

// StateData represents individual elevator state in JSON format
type StateData struct {
	//Id          string                 `json:"id"`
	Behaviour   string   `json:"behaviour"`
	Floor       int                      `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool                   `json:"cabRequests"`
}

// ElevatorMessage is the complete JSON message structure for hall request assigner input
type ElevatorMessage struct {
	HallRequests [][]bool             `json:"hallRequests"`
	States       map[string]StateData `json:"states"`
}

// ConvertToJson converts elevator states to JSON format for hall request assigner.
// It takes the current elevator states and hall requests, structures them according to the expected format, and returns the JSON string.
func ConvertToJson(myId string,
	cabOrders map[string]*orders.CabOrders,
	hallOrders *orders.HallOrders,
	elevators map[string]*elevator.Elevator) (string, error) {

	hallRequestsArray := hallOrders.Simplify()
	// Convert elevator states and cab orders into the expected JSON format
	states := make(map[string]StateData)
	for elevID, elev := range elevators {
		states[elevID] = StateData{
			//Id:          elevID,
			Behaviour:   elev.BehaviourString(),
			Floor:       elev.CurrentFloor(),
			Direction:   elev.DirectionString(),
			CabRequests: cabOrders[elevID].Simplify(),
		}
	}

	message := ElevatorMessage{
		HallRequests: hallRequestsArray,
		States:       states,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	return string(jsonData), nil
}
