package main

import "Driver-go/elevio"

type Behaviour string

const (
	BehaviourIdle     Behaviour = "idle"
	BehaviourMoving   Behaviour = "moving"
	BehaviourDoorOpen Behaviour = "doorOpen"
)

type Direction string

const (
	DirectionUp   Direction = "up"
	DirectionDown Direction = "down"
	DirectionStop Direction = "stop"
)

type State struct {
	Behaviour   Behaviour `json:"behaviour"`
	Floor       uint      `json:"floor"` // NonNegativeInteger
	Direction   Direction `json:"direction"`
	CabRequests []bool    `json:"cabRequests"`
}

type Payload struct {
	HallRequests [][]bool         `json:"hallRequests"`
	States       map[string]State `json:"states"`
}

func InitSingleElevator(numFloors int) (
	chan elevio.ButtonEvent,
	chan int,
	chan bool,
	chan bool,
) {

	elevio.Init("localhost:15657", numFloors)

	// var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	return drv_buttons, drv_floors, drv_obstr, drv_stop
}
