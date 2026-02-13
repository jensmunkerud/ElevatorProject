package elevatorstruct

type Elevator struct {
	ID       int
	Behaviour string
	Floor    int
	Direction string
	cabRequest []int
}