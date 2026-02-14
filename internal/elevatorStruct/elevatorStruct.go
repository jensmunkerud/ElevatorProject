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

func (e *Elevator) initialize(id int) {
	e.id = id //[up, down] * numFloors
	e.behaviour = "idle"
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = make([]bool, numFloors)
}