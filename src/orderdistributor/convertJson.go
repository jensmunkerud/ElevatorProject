package orderdistributor

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"encoding/json"
	"fmt"
)

// StateData represents individual elevator state in JSON format
type StateData struct {
	//Id          string                 `json:"id"`
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

// ElevatorMessage is the complete JSON message structure for hall request assigner input
type ElevatorMessage struct {
	HallRequests [][]bool             `json:"hallRequests"`
	States       map[string]StateData `json:"states"`
}

// ConvertToJson converts elevator states to JSON format for hall request assigner.
// It takes the current elevator states and hall requests, structures them according to the expected format, and returns the JSON string.
func ConvertToJson(cabOrders map[string]*orders.CabOrders,
	hallOrders *orders.HallOrders,
	elevators map[string]*elevator.Elevator) (string, error) {

	hallRequestsArray := hallOrders.Simplify()
	// Convert elevator states and cab orders into the expected JSON format
	states := make(map[string]StateData)
	for elevID, elev := range elevators {
		var cabReqs []bool
		if cab, ok := cabOrders[elevID]; ok && cab != nil {
			cabReqs = cab.Simplify()
		} else {
			cabReqs = make([]bool, config.NumFloors)
		}
		states[elevID] = StateData{
			Behaviour:   elev.BehaviourToString(),
			Floor:       elev.CurrentFloor(),
			Direction:   elev.DirectionString(),
			CabRequests: cabReqs,
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

// A function that converts the JSON output from the cost function into a map of elevator IDs.
// Each elevator ID maps to a struct containing the button states for that elevator. The JSON is expected to be in the format:
// ID1: [[upButtonState, downButtonState], [upButtonState, downButtonState], ...],
// ID2: [[upButtonState, downButtonState], [upButtonState, downButtonState], ...], ...
func ConvertFromJson(jsonStr string) (map[string][config.NumFloors][config.NumButtons]bool, error) {
	var results map[string][config.NumFloors][config.NumButtons]bool

	err := json.Unmarshal([]byte(jsonStr), &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
