package elevatorstruct

const (
	numFloors  = 4
	numButtons = 3
)

type Elevator struct {
	HallRequests [numFloors][2]bool //[up, down]
	id           int
	behaviour    string
	floor        int
	direction    string
	cabRequest   []bool
}

func (e *Elevator) Initialize(id int, currentFloor int, direction string) {
	e.HallRequests = [numFloors][2]bool{} //[up, down] * numFloors
	e.id = id                             //[up, down] * numFloors
	e.behaviour = "idle"
	e.floor = currentFloor
	e.direction = direction
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = make([]bool, numFloors)
}

func (e *Elevator) CurrentElevatorId() int {
	return e.id
}

func (e *Elevator) Behaviour() string {
	return e.behaviour
}

func (e *Elevator) Floor() int {
	return e.floor
}

func (e *Elevator) Direction() string {
	return e.direction
}

func (e *Elevator) CabRequests() []bool {
	return e.cabRequest
}
