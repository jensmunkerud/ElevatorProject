package elevatorstruct

type Elevator struct {
	numFloors int
	hallResquests [][2]bool //[up, down]
	ID       int
	Behaviour string
	Floor    int
	Direction string
	cabRequest []bool

}

func (e *Elevator) initialize(numFloors int, ID int) {
	e.numFloors = numFloors
	e.ID = ID
	e.hallResquests = make([][2]bool, numFloors) //[up, down] * numFloors
	e.Behaviour = "idle"
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = make([]bool, numFloors)
}