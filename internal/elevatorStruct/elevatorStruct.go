package elevatorstruct

type Elevator struct {
	numFloors int
	hallResquests [][]bool
	ID       int
	Behaviour string
	Floor    int
	Direction string
	cabRequest []bool

}

func (e Elevator) initialize(numFloors int, ID int) Elevator {
	e.numFloors = numFloors
	e.ID = ID
	e.hallResquests = make([][]bool, 2)
	e.Behaviour = "idle"
	// Check which floor it is in
	// Read what direction it is moving
	e.cabRequest = [][]bool{}
}