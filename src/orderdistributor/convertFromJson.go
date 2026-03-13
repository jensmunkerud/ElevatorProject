package orderdistributor

import (
	"encoding/json"
)

//A function that converts the JSON output from the cost function into a map of elevator IDs.
//Each elevator ID maps to a struct containing the button states for that elevator. The JSON is expected to be in the format:
//ID1: [[upButtonState, downButtonState], [upButtonState, downButtonState], ...],
//ID2: [[upButtonState, downButtonState], [upButtonState, downButtonState], ...], ...

func ConvertFromJson(jsonStr string) (map[string][][]bool, error) {
	var results map[string][][]bool

	err := json.Unmarshal([]byte(jsonStr), &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
