package callhandler


func ParseElevatorJson(jsonStr string) (map[string]ElevatorButtons, error) {
    var rawData map[string][][]bool
    
    err := json.Unmarshal([]byte(jsonStr), &rawData)
    if err != nil {
        return nil, err
    }
    
    result := make(map[string]ElevatorButtons)
    
    for elevatorID, floors := range rawData {
        var buttons ElevatorButtons
        for i, floor := range floors {
            if i < NumFloors && len(floor) >= 2 {
                buttons[i][0] = floor[0]
                buttons[i][1] = floor[1]
            }
        }
        result[elevatorID] = buttons
    }
    
    return result, nil
}