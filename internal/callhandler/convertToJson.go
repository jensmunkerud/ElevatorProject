package callhandler

import (
	"encoding/json"
	"fmt"

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
func ConvertToJson(hallRequests [4][2]bool, elevators map[string]*es.Elevator) (string, error) {
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

// ConvertToJsonPretty converts elevator states to formatted JSON
func ConvertToJsonPretty(hallRequests [4][2]bool, elevators map[string]*es.Elevator) (string, error) {
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

	// Marshal to formatted JSON
	jsonData, err := json.MarshalIndent(message, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	return string(jsonData), nil
}

// ParseAssignerOutput parses the hall request assigner JSON output
func ParseAssignerOutput(jsonOutput string) (AssignerOutput, error) {
	var output AssignerOutput
	err := json.Unmarshal([]byte(jsonOutput), &output)
	if err != nil {
		return nil, fmt.Errorf("error parsing assigner output: %w", err)
	}
	return output, nil
}

// UpdateHallRequests updates elevator hall requests from assigner output
func UpdateHallRequests(elevators map[string]*es.Elevator, assignerOutput AssignerOutput) error {
	for id, elev := range elevators {
		if hallRequests, exists := assignerOutput[id]; exists {
			if len(hallRequests) != 4 {
				return fmt.Errorf("invalid hall requests length for elevator %s: expected 4, got %d", id, len(hallRequests))
			}
			for floor := 0; floor < 4; floor++ {
				if len(hallRequests[floor]) != 2 {
					return fmt.Errorf("invalid floor requests length for elevator %s floor %d: expected 2, got %d", id, floor, len(hallRequests[floor]))
				}
				elev.HallRequests[floor][0] = hallRequests[floor][0]
				elev.HallRequests[floor][1] = hallRequests[floor][1]
			}
		}
	}
	return nil
}
