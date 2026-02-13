package main

import (
	"fmt"
	"Driver-go"
)

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
