package elevatorstruct

const (
	numFloors = 4
	numButtons = 3
)

type Elevator struct {
	HallRequests [numFloors][2]bool //[up, down]
	id       int
	behaviour string
	floor    int
	Direction string
	cabRequest []bool

}

func (e *Elevator) Initialize(id int, currentFloor int, direction string) {
	e.HallRequests = [numFloors][2]bool{} //[up, down] * numFloors
	e.id = id //[up, down] * numFloors
	e.behaviour = "idle"
	e.floor = currentFloor
	e.Direction = direction
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = make([]bool, numFloors)
}