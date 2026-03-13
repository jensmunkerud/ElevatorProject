package config

import "time"

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3
	Timeout           = 1000 * time.Millisecond
	HeartbeatInterval = 50 * time.Millisecond
)

var (
	myID = "placeholder"
)