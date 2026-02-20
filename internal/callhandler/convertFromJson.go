package callhandler

import (
	"elevatorproject/internal/config"
    "elevatorproject/internal/elevatorStruct"
	"encoding/json"
)

func ParseElevatorJson(jsonStr string) (map[string]elevatorstruct.ElevatorButtons, error) {
	var rawData map[string][][]bool

	err := json.Unmarshal([]byte(jsonStr), &rawData)
	if err != nil {
		return nil, err
	}

	result := make(map[string]elevatorstruct.ElevatorButtons)

	for elevatorID, floors := range rawData {
		var buttons elevatorstruct.ElevatorButtons
		for floorNum, buttonDir := range floors {
			if floorNum < config.NumFloors && len(buttonDir) >= 2 {
				buttons.Buttons[floorNum][0] = buttonDir[0]
				buttons.Buttons[floorNum][1] = buttonDir[1]
			}
		}
		result[elevatorID] = buttons
	}

	return result, nil
}
