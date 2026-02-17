package callhandler

import (
	"encoding/json"
	"fmt"

	"elevatorproject/internal/config"
	es "elevatorproject/internal/elevatorstruct"
)

// StateData represents individual elevator state in JSON format
type StateData struct {
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

// AssignerOutput is the structure for hall request assigner output
// Maps elevator ID to their assigned hall requests
type AssignerOutput map[string][][]bool

// ConvertToJson converts elevator states to JSON format for hall request assigner
func ConvertToJson(hallRequests [config.NumFloors][2]bool, elevators map[string]*es.Elevator) (string, error) {
	// Convert hall requests from [4][2]bool to [][]bool
	hallRequestsArray := make([][]bool, len(hallRequests))
	for i, floor := range hallRequests {
		hallRequestsArray[i] = []bool{floor[0], floor[1]}
	}

	// Convert elevator states
	states := make(map[string]StateData)
	for id, elev := range elevators {
		states[id] = StateData{
			Behaviour:   elev.Behaviour(),
			Floor:       elev.Floor(),
			Direction:   elev.Direction(),
			CabRequests: elev.CabRequests(),
		}
	}

	// Create the message
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
