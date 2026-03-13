package config

import "time"

const (
	NumFloors         = 4
	NumButtons        = 3
	NumElevators      = 3
	DoorOpenDuration  = 3000 * time.Millisecond
	Timeout           = 1000 * time.Millisecond
	HeartbeatInterval = 50 * time.Millisecond
)

var (
	// MyID is the local elevator's ID (set by the application at startup).
	// It is used by packages that need to select "my" state from shared maps.
	MyID = "placeholder"
)