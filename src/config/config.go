package config

import "time"

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3
	Timeout           = 1000 * time.Millisecond
	HeartbeatInterval = 50 * time.Millisecond
	PeersPort = 15647
	BroadcastPort = 16569
)

var (
	MyID = "placeholder"
)