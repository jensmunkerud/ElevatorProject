package config

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	NumFloors    = 4
	NumButtons   = 3
	NumElevators = 3

	DoorOpenDuration        = 3000 * time.Millisecond
	ServiceTimeout          = 4000 * time.Millisecond
	Timeout                 = 1000 * time.Millisecond
	HeartbeatInterval       = 100 * time.Millisecond
	ProcessPairRestartDelay = 10 * time.Second

	ElevatorPort  = 15657
	PeersPort     = 52317
	BroadcastPort = 53491

	testing = true
)

var (
	// ElevatorIOPort is the address for elevator I/O. In testing mode it is set from user input in SetMyID.
	ElevatorIOPort = "localhost:15658"
	myID           = "placeholder"
	once           sync.Once
)

// MyID returns the local elevator ID (set once from MAC address or via SetMyID).
func MyID() string {
	return myID
}

// SetMyID initializes config with the default port. Used by tests.
func SetMyID() {
	InitConfig(15658)
}

// InitConfig sets the local elevator ID and ElevatorIOPort once.
// In testing mode: ID = port, ElevatorIOPort = localhost:port.
// In production mode: ID = MAC address, ElevatorIOPort = default.
func InitConfig(port int) {
	once.Do(func() {
		if testing {
			myID = strconv.Itoa(port)
			ElevatorIOPort = "localhost:" + strconv.Itoa(port)
		} else {
			mac, err := getMacAddr()
			attempts := 0
			for err != nil {
				fmt.Printf("Error: %v, retrying MAC address...\n", err)
				mac, err = getMacAddr()
				time.Sleep(1 * time.Second)
				attempts++
				if attempts > 10 {
					panic("Failed to find MAC address after 10 attempts")
				}
			}
			myID = mac
		}
	})
}

func getMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a, nil
		}
	}
	return "", fmt.Errorf("no MAC address found")
}
