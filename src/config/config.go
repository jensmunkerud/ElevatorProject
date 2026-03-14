package config

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	NumFloors         = 4
	NumButtons        = 3
	NumElevators      = 3
	DoorOpenDuration  = 3000 * time.Millisecond
	Timeout           = 1000 * time.Millisecond
	HeartbeatInterval = 50 * time.Millisecond
	PeersPort = 15647
	BroadcastPort = 16569
)

var (
	myID = "placeholder"
	once sync.Once
)

// MyID returns the local elevator ID (set once from MAC address or via SetMyID).
func MyID() string {
	return myID
}

// SetMyID sets the local ID once. If id is empty, myID is set from the machine's MAC address.
// Only the first call has effect. The application should call SetMyID("") at startup to use MAC-based ID.
func SetMyID() {
	once.Do(func() {
		mac, err := getMacAddr()
		attempts := 0
		for err != nil {
			fmt.Printf("Error: %v, retrying MAC address...\n", err)
			mac, err = getMacAddr() // keep "placeholder" on error
			time.Sleep(1 * time.Second)
			attempts++
			
			if attempts > 10 {
				panic("Failed to find MAC address after 10 attempts")
			}
		}
		myID = mac
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


